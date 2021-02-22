package bulkprocess

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
)

func TestListFromGoogleStorage(t *testing.T) {

	sclient, err := storage.NewClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	files, err := ListFromGoogleStorage("gs://gcp-public-data-landsat/LC08/01/044/034/LC08_L1GT_044034_20130330_20170310_01_T2", sclient)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(files)
}
