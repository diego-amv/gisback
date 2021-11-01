package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"log"
)

func (t *GISMessage) keyLocal() *datastore.Key {
	if t.Id == 0 {
		return datastore.IncompleteKey( "GIS", nil)
	}
	return datastore.IDKey("GIS",  t.Id, nil)
}

func (t *GISMessage) saveLocal(c *datastore.Client, ctx context.Context) (*GISMessage, error) {
	k, err := c.Put(ctx, t.keyLocal(), t)
	if err != nil {
		return nil, err
	}
	t.Id = k.ID
	return t, nil
}

func getAllGISLocal(c *datastore.Client, ctx context.Context, offset int) ([]GISMessage, error) {
	var data []GISMessage
	q := datastore.NewQuery("GIS").Order("-Time").Limit(5).Offset(offset)
	if _, err := c.GetAll(ctx, q, &data); err != nil {
		log.Printf("Error en obtención de todos los registros")
		return nil, err
	}
	return data, nil
}

func searchGISLocal(c *datastore.Client, ctx context.Context, imei string, offset int) ([]GISMessage, error) {
	var data []GISMessage
	q := datastore.NewQuery("GIS").Limit(5).Offset(offset).Order("-Time").Filter("IMEI=", imei)
	if _, err := c.GetAll(ctx, q, &data); err != nil {
		log.Printf("Error en obtención de registros")
		return nil, err
	}
	return data, nil
}

func countLocal(c *datastore.Client, ctx context.Context, imei int) int {
	var q *datastore.Query
	if imei == 0 {
		q = datastore.NewQuery("GIS")
	} else {
		q = datastore.NewQuery("GIS").KeysOnly().Filter("IMEI=", imei)
	}
	total, err := c.Count(ctx, q)
	if err != nil {
		log.Printf("Error en obtención de todos los registros")
		return 0
	}
	return total
}