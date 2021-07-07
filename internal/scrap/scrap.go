package scrap

import (
	"context"
	"errors"
	"log"
	"net/http"

	env "github.com/Netflix/go-env"
	"github.com/PuerkitoBio/goquery"
	jsoniter "github.com/json-iterator/go"
	"github.com/yoannduc/scoreav-crawler/internal/helpers"
	"github.com/yoannduc/scoreav-crawler/internal/types"
)

type config struct {
	DdbTable string `env:"AWS_DDB_TABLE_NAME,required=true"`
	DdbSave  bool   `env:"SAVE_TO_DDB,default=true"`
}

type output struct {
	types.Event
	Processed int `json:"processed"`
}

func buildURI(e types.Event) (string, error) {
	switch e.Source {
	case "scoreav":
		switch e.Type {
		case "news":
			return "http://www.scoreav.com/news/", nil
		case "focus":
			return "http://www.scoreav.com/chronique/", nil
		default:
			return "", errors.New("Type empty or not handled for source " + e.Source + " (inputed: " + e.Type + ")")
		}
	default:
		return "", errors.New("Source not handled or empty (inputed:\"" + e.Source + "\")")
	}
}

// TODO Add some logs
// TODO Think about db keys with uniqueness in mind

// HandleRequest handles the lambda job
// It scraps scoreav site to get articles of today which are not yet in ddb
// ctx is the lambda context
// e is the event input
func HandleRequest(ctx context.Context, e types.Event) (string, error) {
	// Set config
	var c config
	if _, err := env.UnmarshalFromEnviron(&c); err != nil {
		return helpers.HandleErr("Error while unwrapping env config : " + err.Error())
	}

	uri, err := buildURI(e)
	if err != nil {
		return helpers.HandleErr("Error building the uri: " + err.Error())
	}

	resp, err := http.Get(uri)
	if err != nil {
		return helpers.HandleErr("Error while querying the endpoint :" + err.Error() + " (endpoint: " + uri + ")")
	}
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return helpers.HandleErr("Error while decoding response into goquery: " + err.Error())
	}

	els := buildElemList(d, e)

	// If we do not want to save to db (to debug scrapping), return scrapped list
	if !c.DdbSave {
		j, err := jsoniter.MarshalToString(els)
		if err != nil {
			return helpers.HandleErr(err.Error())
		}

		log.Println(j)

		return j, nil
	}

	err = writeToDdb(els, c.DdbTable)
	if err != nil {
		return helpers.HandleErr(err.Error())
	}

	j, err := jsoniter.MarshalToString(&output{
		Processed: len(els),
		Event:     e,
	})
	if err != nil {
		return helpers.HandleErr(err.Error())
	}

	log.Println(j)

	return j, nil
}
