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
	"slices"
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
	"github.com/wcharczuk/go-chart"
)

var (
	companyName = "Ingenieros AJV"
	companyAddress = "Calle tal #tal-tal frente a tal"
	companyPhone = "3051234567"
	tmpl = template.Must(template.ParseFiles("resources/report_template/template.html"))
	threshold = 50.0
)

type ReportsService struct {
	projectsRepo project.Repository
}

type Report struct {
    Filename string
    File []byte
}

func NewReportsService (repo project.Repository) *ReportsService {
	return &ReportsService{projectsRepo: repo}
}

func (r *ReportsService) GenerateReportForOneFamily(projectID int, familyID int) (*Report, error) {
	project, err := r.projectsRepo.GetProjectByID(projectID)

	if err != nil {
		return nil, err
	}

	var family *family.Family
	for _, p := range(project.Families) {
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

func (r *ReportsService) generateReportsChart(members []member.Member) string {
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
        Name: "Threshold",
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
			Name string
			Address string
			Phone string
		}
		Client struct {
			Name string
			ID int
		}
		Project struct {
			Name string
			ReportDate string
		}
		Family struct {
			Name string
		}
		Members []struct {
			ID int
			FracturedAt string
			Result int
			Operative string
		}
		ChartBase64 string
	}{}

	data.Company.Name = companyName
	data.Company.Address = companyAddress
	data.Company.Phone = companyPhone
	data.Client.Name = project.Client.Name
	data.Client.ID = project.ClientID
	data.Project.Name = project.Name
	data.Project.ReportDate = time.Now().Format("2006-01-02 15:04:05")
	data.Family.Name = "Ubicación donde fue tomada la muestra"
	for _, v := range(family.Members) {
		if v.IsReported != nil && *v.IsReported {
			log.Print(v.FracturedAt)
			data.Members = append(data.Members, struct{ID int; 
			FracturedAt string; 
			Result int; 
			Operative string}{
				ID: v.ID, 
				FracturedAt: v.FracturedAt.Local().Format("2006-01-02 15:04:05"), 
				Result: *v.Result,
				Operative: fmt.Sprintf("%s %s", v.Operative.FirstName, v.Operative.LastName)})
		}
		
	}
	data.ChartBase64 = r.generateReportsChart(family.Members)
	return data
}

func htmlToPDFWithWK(html []byte, outputPath string) error {
    cmd := exec.Command("C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe",
        "--enable-local-file-access",
        "--encoding", "utf-8",
        "--margin-top", "15mm",
        "--margin-bottom", "15mm",
        "-",
        outputPath, // "-" = read from stdin
    )

    cmd.Stdin = bytes.NewReader(html)

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("wkhtmltopdf error: %v - %s", err, stderr.String())
    }

    return nil
}

