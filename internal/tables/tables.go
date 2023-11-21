package tables

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SpreadsheetService struct {
	spreadsheetID string
	service       *sheets.Service
}

func NewSpreadsheetService(gCloudConfPath, spreadsheetID string) (*SpreadsheetService, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(
		ctx,
		option.WithCredentialsFile(gCloudConfPath),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		return nil, err
	}
	defer ctx.Done()

	c := &SpreadsheetService{
		spreadsheetID: spreadsheetID,
		service:       srv,
	}

	return c, nil
}

func (s *SpreadsheetService) Test() error {
	_, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, "A1:Z1").Do()
	return err
}

func (s *SpreadsheetService) Read(
	startCell, endCell string,
) (*sheets.ValueRange, error) {
	values, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, fmt.Sprintf("%s:%s", startCell, endCell)).
		Do()
	return values, err
}

func (s *SpreadsheetService) Write(startCell string, endCell string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}
	_, err := s.service.Spreadsheets.Values.Update(s.spreadsheetID, fmt.Sprintf("%s:%s", startCell, endCell), valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	return err
}
