package main

import (
	"context"
	"google.golang.org/appengine/v2/datastore"
)

func defaultGIS(c context.Context) *datastore.Key {
	return datastore.NewKey(c, "GIS", "default", 0, nil)
}

func (t *GISMessage) key(c context.Context) *datastore.Key {
	if t.Id == 0 {
		return datastore.NewIncompleteKey(c, "GIS", defaultGIS(c))
	}
	return datastore.NewKey(c, "GIS", "", t.Id, defaultGIS(c))
}

func (t *GISMessage) save(c context.Context) (*GISMessage, error) {
	k, err := datastore.Put(c, t.key(c), t)
	if err != nil {
		return nil, err
	}
	t.Id = k.IntID()
	return t, nil
}

func getAllGIS(c context.Context, offset int) ([]GISMessage, error) {
	var data []GISMessage
	ks, err := datastore.NewQuery("GIS").Limit(5).Offset(offset).Ancestor(defaultGIS(c)).Order("-Time").GetAll(c, &data)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(data); i++ {
		data[i].Id = ks[i].IntID()
	}
	return data, nil
}

func searchGIS(c context.Context, imei string, offset int) ([]GISMessage, error) {
	var data []GISMessage
	ks, err := datastore.NewQuery("GIS").Limit(5).Offset(offset).KeysOnly().Ancestor(defaultGIS(c)).Order("-Time").Filter("IMEI=", imei).GetAll(c, nil)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(data); i++ {
		data[i].Id = ks[i].IntID()
	}
	return data, nil
}

func count(ctx context.Context, imei int) int {
	var total int
	if imei == 0 {
		total, _ = datastore.NewQuery("GIS").Count(ctx)
	} else {
		total, _ = datastore.NewQuery("GIS").KeysOnly().Filter("IMEI=", imei).Count(ctx)
	}
	return total
}
