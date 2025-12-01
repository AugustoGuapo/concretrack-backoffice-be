package application

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
	"github.com/wcharczuk/go-chart"
)

var (
	companyName    = "Ingenieros AJV"
	companyAddress = "Calle tal #tal-tal frente a tal"
	companyPhone   = "3051234567"
	tmpl           = template.Must(template.ParseFiles(templatePath()))
)

func templatePath() string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Dir(file)
	return filepath.Join(base, "..", "..", "resources", "report_template", "template.html")
}

type ReportsService struct {
	projectsRepo project.Repository
}

type Report struct {
	Filename string
	File     []byte
}

type ReportMember struct {
	ID               int
	FracturedAt      string
	Result           float64
	Operative        string
	SamplePlace      string
	DateOfEntry      string
	AgeDays          int
	DiameterCM       float64
	LengthCM         float64
	AreaCM2          string
	AdjustmentFactor int
	StrengthKGCM2    string
	StrengthPSI      string
	DesignMPA        string
	DesignPSI        string
	ObtainedPercent  string
	FailureShape     string
    Perpendicularity string
}

func NewReportsService(repo project.Repository) *ReportsService {
	return &ReportsService{projectsRepo: repo}
}

func (r *ReportsService) GenerateReportForOneFamily(projectID int, familyID int) (*Report, error) {
	project, err := r.projectsRepo.GetProjectByID(projectID)

	if err != nil {
		return nil, err
	}

	var family *family.Family
	for _, p := range project.Families {
		if familyID == p.ID {
			family = &p
			break
		}
	}

	data := r.generateReportData(project, family)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)

	if err != nil {
		return nil, err
	}

	html := buf.Bytes()
	tmp, err := os.CreateTemp("", "reporte-*.pdf")
	if err != nil {
		return nil, err
	}
	outputPath := tmp.Name()
	tmp.Close()
	defer os.Remove(outputPath)
	if err := htmlToPDFWithWK(html, outputPath); err != nil {
		return nil, err
	}
	pdfBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("Reporte %v-%v-%v.pdf", project.Name, "Lugar de toma", time.Now().Format("2006-01-02 15:04:05"))

	return &Report{Filename: filename, File: pdfBytes}, nil
}

func (r *ReportsService) generateReportsChart(members []member.Member, threshold float64) string {
	// 1. Preparar datos
	var xValues []float64
	var yValues []float64

	for i, v := range members {
		// Para X usamos un índice o timestamp — go-chart NO toma strings en X
		xValues = append(xValues, float64(i))
		yValues = append(yValues, float64(*v.Result))
	}
	if len(yValues) < 1 || len(xValues) < 1 {
		return ""
	}

	maxY := math.Max(threshold, slices.Max(yValues))

	// 2. Serie principal
	dataSeries := chart.ContinuousSeries{
		Name:    "Resultados",
		XValues: xValues,
		YValues: yValues,
		Style: chart.Style{
			Show:        true,
			StrokeWidth: 3,
		},
	}

	// 3. Threshold (línea horizontal)
	thresholdSeries := chart.ContinuousSeries{
		Name:    "Threshold",
		XValues: []float64{xValues[0], xValues[len(xValues)-1]},
		YValues: []float64{threshold, threshold},
		Style: chart.Style{
			Show:            true,
			StrokeWidth:     2,
			StrokeDashArray: []float64{5, 5},
			StrokeColor:     chart.ColorAlternateGray,
		},
	}

	// 4. Configurar gráfica con cuadrícula, ejes, paddings, labels, título…
	graph := chart.Chart{
		Width:  1280,
		Height: 720,
		Title:  "Resultados vs Tiempo",

		TitleStyle: chart.Style{
			Show:        true,
			FontSize:    20,
			StrokeColor: chart.ColorBlack,
		},

		Background: chart.Style{
			Padding: chart.Box{
				Top:  40,
				Left: 60,
			},
		},

		XAxis: chart.XAxis{
			Name:      "Fecha de fractura",
			NameStyle: chart.Style{Show: true, FontSize: 14},
			Style:     chart.Style{Show: true},

			// Ticks: usar los nombres reales de las fechas
			ValueFormatter: func(v interface{}) string {
				idx := int(v.(float64))
				if idx < 0 || idx >= len(members) {
					return ""
				}
				return members[idx].FracturedAt.Format("2006-01-02")
			},

			GridMajorStyle: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 0.5,
			},
		},

		YAxis: chart.YAxis{
			Name:      "Resistencia (kg/cm²)",
			NameStyle: chart.Style{Show: true, FontSize: 14},
			Style:     chart.Style{Show: true},

			Range: &chart.ContinuousRange{
				Min: 0,
				Max: maxY + 10,
			},

			GridMajorStyle: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 0.5,
			},
		},

		Series: []chart.Series{
			dataSeries,
			thresholdSeries,
		},
	}

	// 5. Renderizar en base64
	buf := bytes.NewBuffer([]byte{})
	if err := graph.Render(chart.PNG, buf); err != nil {
		log.Printf("[generateReportsChart] Error: %v", err)
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded
}

func (r *ReportsService) generateReportData(project *project.Project, family *family.Family) interface{} {
	data := struct {
		Company struct {
			Name    string
			Address string
			Phone   string
		}
		Client struct {
			Name string
			ID   int
		}
		Project struct {
			Name       string
			ReportDate string
		}
		Family struct {
			Name             string
			DesignResistance float64
			DateOfEntry      *time.Time
		}
		Members     []ReportMember
		ChartBase64 string
	}{}

	data.Company.Name = companyName
	data.Company.Address = companyAddress
	data.Company.Phone = companyPhone
	data.Client.Name = project.Client.Name
	data.Client.ID = project.ClientID
	data.Project.Name = project.Name
	data.Project.ReportDate = time.Now().Format("2006-01-02 15:04:05")
	data.Family.Name = family.SamplePlace
	for _, v := range family.Members {
		if v.IsReported != nil && *v.IsReported {
			cilynderArea := math.Pi * math.Pow(family.Radius, 2)
			StrengthKGCM2 := (*v.Result / cilynderArea) * 1 * 102
			StrengthPSI := StrengthKGCM2 / 0.07
			data.Members = append(data.Members, ReportMember{
				SamplePlace:      family.SamplePlace,
				DateOfEntry:      v.DateOfFracture.AddDate(0, 0, -*v.FractureDays).Format(("2006-01-02")),
				AgeDays:          *v.FractureDays,
				DiameterCM:       family.Radius * 2,
				LengthCM:         family.Height,
				AreaCM2:          fmt.Sprintf("%.2f", cilynderArea),
				AdjustmentFactor: 1,
				StrengthKGCM2:    fmt.Sprintf("%.2f", StrengthKGCM2),
				StrengthPSI:      fmt.Sprintf("%.2f", StrengthPSI),
				DesignMPA:        fmt.Sprintf("%.2f", family.DesignResistance / 145.0377),
				DesignPSI:        fmt.Sprintf("%.2f", family.DesignResistance),
				ObtainedPercent:  fmt.Sprintf("%.2f", (StrengthPSI / family.DesignResistance) * 100),
				FailureShape:     *v.FractureType,
				ID:               v.ID,
				FracturedAt:      v.FracturedAt.Local().Format("2006-01-02"),
				Result:           *v.Result,
				Operative:        fmt.Sprintf("%s %s", v.Operative.FirstName, v.Operative.LastName),
                Perpendicularity: "Si        No",})
		}

	}
    
	data.ChartBase64 = r.generateReportsChart(family.Members, family.DesignResistance)
	return data
}

func htmlToPDFWithWK(html []byte, outputPath string) error {
	cmd := exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--encoding", "utf-8",

		// Orientación horizontal
		"--orientation", "Landscape",

		// Tamaño de página (opcional pero recomendado)
		"--page-size", "A4",

		// Márgenes
		"--margin-top", "15mm",
		"--margin-bottom", "15mm",
		"--margin-left", "10mm",
		"--margin-right", "10mm",

		"-",        // leer HTML desde stdin
		outputPath, // escribir PDF a archivo
	)

	cmd.Stdin = bytes.NewReader(html)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wkhtmltopdf error: %v - %s", err, stderr.String())
	}

	return nil
}
