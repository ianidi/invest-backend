package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/thecodingmachine/gotenberg-go-client/v7"
)

func GeneratePDF(content string) ([]byte, error) {
	var err error

	httpClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	client := &gotenberg.Client{Hostname: viper.GetString("pdf_host"), HTTPClient: httpClient}

	index, err := gotenberg.NewDocumentFromString("index.html", "<html><head><style>body {padding: 40px;}</style></head><body><h1>Contract</h1>"+content+"</body></html>")
	if err != nil {
		return nil, errors.New("PDF_GENERATION_ERROR")
	}

	// header, _ := gotenberg.NewDocumentFromPath("header.html", "/path/to/file")
	// footer, _ := gotenberg.NewDocumentFromPath("footer.html", "/path/to/file")
	// style, _ := gotenberg.NewDocumentFromPath("style.css", "/path/to/file")
	// img, _ := gotenberg.NewDocumentFromPath("img.png", "/path/to/file")

	req := gotenberg.NewHTMLRequest(index)
	// req.Header(header)
	// req.Footer(footer)
	// req.Assets(style) //, img
	req.PaperSize(gotenberg.A4)
	req.Margins(gotenberg.NoMargins)
	req.Scale(0.75)

	filepath := "/var/server/static/" + uuid.New().String() + ".pdf"

	err = client.Store(req, filepath)
	if err != nil {
		return nil, errors.New("PDF_WRITE_ERROR")
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("PDF_READ_ERROR")
	}

	err = os.Remove(filepath)
	if err != nil {
		return nil, errors.New("PDF_REMOVE_ERROR")
	}

	return data, nil
}
