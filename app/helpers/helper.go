package helpers

import (
	"backendtku/app/models"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	_ "github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"

	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"

	_ "github.com/johnfercher/maroto/v2/pkg/components/code"
	_ "github.com/johnfercher/maroto/v2/pkg/components/col"
	_ "github.com/johnfercher/maroto/v2/pkg/core/entity"

	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"

	"gopkg.in/gomail.v2"
)

func ResponseJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "tkflumum@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "tkflumum@gmail.com", "koun gfmt amsc dfci")

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func GetTypeBase64(data string) (feedback string) {
	switch data[0:4] {
	case "/9j/":
		feedback = ".jpg"
	case "iVBO":
		feedback = ".png"
	case "R0lG":
		feedback = ".gif"
	case "JVBE":
		feedback = ".pdf"
	default:
		feedback = ".zip"
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetMarotoAbror(id string, brand string, car string, plat string, name string, cont string, sdate string, edate string, price int, results []string, chassis string, engine string, img1 string, img2 string, img3 string, img4 string) core.Maroto {
	customFont := "Muli"
	customFonts, errFont := repository.New().
		AddUTF8Font(customFont, fontstyle.Normal, "upload/policy/Muli.ttf").
		AddUTF8Font(customFont, fontstyle.Bold, "upload/policy/Muli-Bold.ttf").
		Load()
	if errFont != nil {
		log.Fatal(errFont.Error())
	}

	bytes, errImg := os.ReadFile("upload/policy/watermark2.png")
	if errImg != nil {
		log.Fatal(errImg)
	}
	b := config.NewBuilder().
		WithLeftMargin(10).
		WithRightMargin(10).
		WithTopMargin(0).
		WithCustomFonts(customFonts).
		WithDefaultFont(&props.Font{Family: customFont}).
		WithBackgroundImage(bytes, extension.Png).
		WithPageSize(pagesize.A4)

	mrt := maroto.New(b.Build())
	m := maroto.NewMetricsDecorator(mrt)

	errHeader := m.RegisterHeader(
		row.New(23).Add(
			image.NewFromFileCol(4, "upload/policy/header1.png"),
			image.NewFromFileCol(6, "upload/policy/header2.png", props.Rect{Percent: 80, Top: 3, Left: 20})),
		row.New(10).Add(text.NewCol(12, "Polis Takaful Abror "+brand+" "+car+" "+plat, props.Text{Align: align.Center, Top: 5, Style: fontstyle.Bold, Size: 12})))

	if errHeader != nil {
		log.Fatal(errHeader)
	}

	errFooter := m.RegisterFooter(
		row.New(23).Add(
			text.NewCol(4, "Kantor Pusat                                      Jl. Persada Raya No. 70 C-D Menteng Dalam, Tebet,           Jakarta, 1287", props.Text{Size: 9, Top: 3, Left: 6}),
			image.NewFromFileCol(3, "upload/policy/footer1.png", props.Rect{Percent: 80, Top: 3}).Add(
				text.New("Telp : (021) 285 43 111", props.Text{Top: 9, Left: 1, Size: 9}),
				text.New("www.takafulumum.co.id", props.Text{Top: 12, Left: 1, Size: 9}),
			),
			image.NewFromFileCol(3, "upload/policy/footer2.png", props.Rect{Percent: 80, Left: 10}),
			image.NewFromFileCol(3, "upload/policy/footer3.png", props.Rect{Percent: 65, Left: 5, Top: 2})),
	)

	if errFooter != nil {
		log.Fatal(errFooter)
	}

	m.AddRows(text.NewRow(10, "Polis. : "+id, props.Text{Size: 10, Top: 5}))

	colStyle := &props.Cell{
		BorderType: border.None,
	}

	var userData = [][]string{
		{"1. Nama Pemegang Polis", "8", name},
		{"3. Periode Perjalanan", "8", sdate + "  -  " + edate},
		{"4. Klausul dan Warranty", "8", "Klausula Takaful Safari"},
		{"5. Risiko Sendiri", "8", cont},
	}

	for i, content := range userData {
		number, _ := strconv.ParseFloat(content[1], 64)
		m.AddRows(row.New(number).Add(
			text.NewCol(4, content[0], props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, content[2], props.Text{Size: 10}).WithStyle(colStyle),
		))
		if i == 0 {
			m.AddRows(row.New(8).Add(
				text.NewCol(4, "2. Luas Jaminan", props.Text{Size: 10}).WithStyle(colStyle),
				text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
				text.NewCol(7, cont, props.Text{Size: 10}).WithStyle(colStyle),
			))

			subset := results[0:12]
			for i, benefit := range subset {
				height := 8
				if i == 0 {
					height = 12
				}
				m.AddRows(row.New(float64(height)).Add(
					text.NewCol(4, " ", props.Text{Size: 10}).WithStyle(colStyle),
					text.NewCol(1, " ", props.Text{Size: 10}).WithStyle(colStyle),
					text.NewCol(7, benefit, props.Text{Size: 10}).WithStyle(colStyle),
				))
			}
		}
	}

	subset := results[12:15]
	for _, benefit := range subset {
		m.AddRows(row.New(8).Add(
			text.NewCol(4, " ", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, " ", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, benefit, props.Text{Size: 10}).WithStyle(colStyle),
		))
	}

	m.AddRows(
		row.New(8).Add(
			text.NewCol(4, "7. Perhitungan Kontribusi", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, FormatMoney(price), props.Text{Size: 10, Align: align.Right}).WithStyle(colStyle),
		),
		row.New(8).Add(
			text.NewCol(4, "   - Biaya Polis dan Materai", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, "0", props.Text{Size: 10, Align: align.Right}).
				WithStyle(&props.Cell{LineStyle: linestyle.Dashed, BorderType: border.Bottom}),
		),
		row.New(8).Add(
			text.NewCol(4, "TOTAL ", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Right: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10, Style: fontstyle.Bold, Top: 3}).WithStyle(colStyle),
			text.NewCol(7, FormatMoney(price), props.Text{Size: 10, Align: align.Right, Style: fontstyle.Bold, Top: 3}).WithStyle(colStyle),
		),
		row.New(20).Add(
			text.NewCol(6, " ", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "Dibuat di Jakarta", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "Pada tanggal ", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "PT Asuransi Takaful Umum", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(25).Add(
			image.NewFromFileCol(2, "upload/policy/ttd.png", props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
		),
		row.New(10).Add(
			text.NewCol(7, "Raihan Fadhlal Aziz", props.Text{Size: 10}).WithStyle(colStyle),
		),
		row.New(10).Add(text.NewCol(12, "Detail Kendaraan", props.Text{Align: align.Center, Top: 5, Style: fontstyle.Bold, Size: 12})),
		row.New(10).Add(
			text.NewCol(7, "Nomor Mesin : "+engine, props.Text{Size: 10}).WithStyle(colStyle),
		),
		row.New(10).Add(
			text.NewCol(7, "Nomor Rangka : "+chassis, props.Text{Size: 10}).WithStyle(colStyle),
		),
		row.New(10).Add(
			text.NewCol(7, "Kondisi Awal	: ", props.Text{Size: 10}).WithStyle(colStyle),
		),
		row.New(50).Add(
			image.NewFromFileCol(3, "upload/enroll/"+img1, props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
			image.NewFromFileCol(3, "upload/enroll/"+img2, props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
			image.NewFromFileCol(3, "upload/enroll/"+img3, props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
			image.NewFromFileCol(3, "upload/enroll/"+img4, props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
		),
	)
	return m
}

func FormatMoney(price int) string {
	strPrice := strconv.Itoa(price)
	n := len(strPrice)

	result := "Rp "
	for i, digit := range strPrice {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(digit)
	}
	result += "-"

	return result
}

func ExtractPlateCode(plate string) string {
	parts := strings.FieldsFunc(plate, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func GetMaroto(id string, prod string, name string, cont string, sdate string, edate string, price int, results []string, others []models.EnrollmentSafari) core.Maroto {
	customFont := "Muli"
	customFonts, errFont := repository.New().
		AddUTF8Font(customFont, fontstyle.Normal, "upload/policy/Muli.ttf").
		AddUTF8Font(customFont, fontstyle.Bold, "upload/policy/Muli-Bold.ttf").
		Load()
	if errFont != nil {
		log.Fatal(errFont.Error())
	}

	bytes, errImg := os.ReadFile("upload/policy/watermark2.png")
	if errImg != nil {
		log.Fatal(errImg)
	}
	b := config.NewBuilder().
		WithLeftMargin(10).
		WithRightMargin(10).
		WithTopMargin(0).
		WithCustomFonts(customFonts).
		WithDefaultFont(&props.Font{Family: customFont}).
		WithBackgroundImage(bytes, extension.Png).
		WithPageSize(pagesize.A4)

	mrt := maroto.New(b.Build())
	m := maroto.NewMetricsDecorator(mrt)

	errHeader := m.RegisterHeader(
		row.New(23).Add(
			image.NewFromFileCol(4, "upload/policy/header1.png"),
			image.NewFromFileCol(6, "upload/policy/header2.png", props.Rect{Percent: 80, Top: 3, Left: 30})),
		row.New(10).Add(text.NewCol(12, "Polis "+prod, props.Text{Align: align.Center, Top: 5, Style: fontstyle.Bold, Size: 12})))

	if errHeader != nil {
		log.Fatal(errHeader)
	}

	errFooter := m.RegisterFooter(
		row.New(23).Add(
			text.NewCol(4, "Kantor Pusat                                      Jl. Persada Raya No. 70 C-D Menteng Dalam, Tebet,           Jakarta, 1287", props.Text{Size: 9, Top: 10, Left: 6}),
			image.NewFromFileCol(3, "upload/policy/footer1.png", props.Rect{Percent: 80, Top: 17}).Add(
				text.New("Telp : (021) 285 43 111", props.Text{Top: 9, Left: 1, Size: 9}),
				text.New("www.takafulumum.co.id", props.Text{Top: 12, Left: 1, Size: 9}),
			),
			image.NewFromFileCol(3, "upload/policy/footer2.png", props.Rect{Percent: 80, Left: 10, Top: 7}),
			image.NewFromFileCol(3, "upload/policy/footer3.png", props.Rect{Percent: 65, Left: 5, Top: 9})),
	)

	if errFooter != nil {
		log.Fatal(errFooter)
	}

	m.AddRows(text.NewRow(10, "Polis. : "+id, props.Text{Size: 10, Top: 5}))

	colStyle := &props.Cell{
		BorderType: border.None,
	}
	tableStyle := &props.Cell{
		BorderType: border.Full,
	}

	var userData = [][]string{
		{"1. Nama Pemegang Polis", "8", name},
		{"3. Periode Perjalanan", "8", sdate + "  -  " + edate},
		{"4. Klausul dan Warranty", "8", "Klausula Takaful Safari"},
		{"5. Risiko Sendiri", "8", "-"},
	}

	for i, content := range userData {
		number, _ := strconv.ParseFloat(content[1], 64)
		m.AddRows(row.New(number).Add(
			text.NewCol(4, content[0], props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, content[2], props.Text{Size: 10}).WithStyle(colStyle),
		))
		if i == 0 {
			m.AddRows(row.New(8).Add(
				text.NewCol(4, "2. Luas Jaminan", props.Text{Size: 10}).WithStyle(colStyle),
				text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
				text.NewCol(7, cont, props.Text{Size: 10}).WithStyle(colStyle),
			))

			for _, benefit := range results {
				m.AddRows(row.New(8).Add(
					text.NewCol(4, " ", props.Text{Size: 10}).WithStyle(colStyle),
					text.NewCol(1, " ", props.Text{Size: 10}).WithStyle(colStyle),
					text.NewCol(7, benefit, props.Text{Size: 10}).WithStyle(colStyle),
				))
			}
		}
	}

	m.AddRows(
		row.New(8).Add(
			text.NewCol(4, "7. Perhitungan Kontribusi", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, FormatMoney(price), props.Text{Size: 10, Align: align.Right}).WithStyle(colStyle),
		),
		row.New(8).Add(
			text.NewCol(4, "   - Biaya Polis dan Materai", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10}).WithStyle(colStyle),
			text.NewCol(7, "0", props.Text{Size: 10, Align: align.Right}).
				WithStyle(&props.Cell{LineStyle: linestyle.Dashed, BorderType: border.Bottom}),
		),
		row.New(8).Add(
			text.NewCol(4, "TOTAL ", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Right: 10}).WithStyle(colStyle),
			text.NewCol(1, ":", props.Text{Size: 10, Style: fontstyle.Bold, Top: 3}).WithStyle(colStyle),
			text.NewCol(7, FormatMoney(price), props.Text{Size: 10, Align: align.Right, Style: fontstyle.Bold, Top: 3}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "Dibuat di Jakarta", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "Pada tanggal ", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(5).Add(
			text.NewCol(6, "PT Asuransi Takaful Umum", props.Text{Size: 10, Top: 5}).WithStyle(colStyle),
		),
		row.New(25).Add(
			image.NewFromFileCol(2, "upload/policy/ttd.png", props.Rect{
				Center:  false,
				Percent: 80,
				Top:     5,
			}),
		),
		row.New(10).Add(
			text.NewCol(7, "Raihan Fadhlal Aziz", props.Text{Size: 10}).WithStyle(colStyle),
		),
	)

	m.AddRows(text.NewRow(25, "", props.Text{Size: 10, Top: 5}))
	m.AddRows(text.NewRow(10, "Peserta Polis", props.Text{Size: 10, Top: 5}))
	m.AddRows(
		row.New(8).Add(
			text.NewCol(1, "No", props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(4, "Id Pendaftaran", props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(4, "Nama", props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(3, "Tanggal Lahir", props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
		))
	for i, other := range others {
		m.AddRows(row.New(8).Add(
			text.NewCol(1, strconv.Itoa(i+1), props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(4, other.EnrollmentId, props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(4, other.Name, props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
			text.NewCol(3, other.Birthdate, props.Text{Size: 10, Left: 5}).WithStyle(tableStyle),
		))
	}

	return m
}
