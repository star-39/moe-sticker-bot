package core

// func searchLine(kw string) []LineStickerQ {

// }
// var linedataIndex bleve.Index

// func initSearch() {
// 	linedataIndex, err := bleve.Open("linedata.bleve")
// 	if err != nil {
// 		log.Infoln("No Bleve index found, creating new one -> linedata.bleve")
// 		initBleveIndex()
// 	}

// }

// func initBleveIndex() {
// 	mapping := bleve.NewIndexMapping()
// 	index, err := bleve.New("example.bleve", mapping)
// 	if err != nil {
// 		log.Fatalln("bleve: err create index", err)
// 	}

// 	batch := index.NewBatch()
// 	lines := queryLineS("QUERY_ALL")
// 	// Set to true if mapped.
// 	nmap := make(map[string]bool)
// 	for i, l := range lines {
// 		_, exist := nmap[l.Tg_title]
// 		if exist {
// 			l.Tg_title = l.Tg_title + strconv.Itoa(i)
// 		}
// 		batch.Index(l.Tg_title, l)
// 		nmap[l.Tg_title] = true
// 	}
// 	err = index.Batch(batch)
// 	if err != nil {
// 		log.Fatalln("bleve: err create index", err)
// 	}
// 	log.Infoln("Bleve index created.")
// 	index.Close()
// 	linedataIndex, _ = bleve.Open("linedata.bleve")
// }

// func searchLDIndex(kw string) []LineStickerQ {
// 	q := bleve.NewMatchQuery(kw)
// 	req := bleve.NewSearchRequest(q)
// 	req.Fields = []string{"*"}
// 	res, err := linedataIndex.Search(req)
// 	if err != nil {
// 		return nil
// 	}
// 	if len(res.Hits) < 1 {
// 		return nil
// 	}

// 	lsqs := []LineStickerQ{}
// 	for i, hit := range res.Hits {
// 		lsq := LineStickerQ{
// 			Line_id: hit.Fields[],
// 		}
// 	}
// }
