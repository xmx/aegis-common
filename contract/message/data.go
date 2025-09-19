package message

type Data[E any] struct {
	Data E `json:"data"`
}
