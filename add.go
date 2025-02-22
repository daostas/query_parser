package qp

func (qp *QueryParams) Add(data interface{}) {
	switch t := data.(type) {
	case string:
		*qp = append(*qp, NewQueryParam(t, "", "", ""))
	case QueryParams:
		*qp = append(*qp, t...)
	case QueryParam:
		*qp = append(*qp, t)
	}
}
