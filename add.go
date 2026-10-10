package qp

func (qp *QueryParams) Add(data any) {
	switch t := data.(type) {
	case string:
		*qp = append(*qp, QueryParam{
			KeyArray: []string{t},
			Type:     "",
			Sign:     "",
			Value:    "",
			OR:       nil,
		})
	case QueryParams:
		*qp = append(*qp, t...)
	case QueryParam:
		*qp = append(*qp, t)
	}
}
