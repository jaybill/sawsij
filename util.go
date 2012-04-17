package sawsij

import(    
    "unicode"
)

func MakeDbName(fieldName string) string{
	runes       := []rune(fieldName)
	copy        := []rune{}
	usrunes     := []rune("_")
	us          := usrunes[0]
	for i:=0 ; i< len(runes) ; i++ {
		if i > 0 && unicode.IsUpper(runes[i]){		    
		    copy = append(copy,us)
		}
		runes[i] = unicode.ToLower(runes[i])
		copy = append(copy,runes[i])
		
	}
	return string(copy)
}

func MakeFieldName(dbName string) string{

	runes       := []rune(dbName)
	copy        := []rune{}
	usrunes     := []rune("_")
	us          := usrunes[0]
	for i:=0 ; i< len(runes) ; i++ {
		if runes[i] != us {
		    if i == 0{
		        runes[i] = unicode.ToUpper(runes[i])
		    }		    
		    copy = append(copy,runes[i])
		} else {
		    runes[i + 1] = unicode.ToUpper(runes[i + 1])
		}
	}
	return string(copy)
}
