package structs

type Content struct {
	b_name  [12]byte //Nombre del archivo o carpeta
	b_inodo int64    //Apuntador al inodo del archivo o carpeta
}

func NewContent() Content { //Constructor de la estructura Content
	var cont Content
	cont.b_inodo = -1
	return cont
}
