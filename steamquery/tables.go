package main

import (
	"context"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

var (
	spreadsheetID = ""
)

type spreadsheetService struct {
	service *sheets.Service
}

func newSpreadsheetService(gCloudConfPath string) (*spreadsheetService, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(gCloudConfPath), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}

	c := &spreadsheetService{
		service: srv,
	}

	return c, nil
}

func (s *spreadsheetService) testConnection() error {
	_, err := s.service.Spreadsheets.Values.Get(spreadsheetID, "A1:Z1").Do()
	return err
}

func (s *spreadsheetService) writeSingleEntryToTable(cell string, values []interface{}) error {
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, values)

	_, err := s.service.Spreadsheets.Values.Update(spreadsheetID, cell, &vr).ValueInputOption("USER_ENTERED").Do()

	return err
}
