package gribble

// import ( 
// "fmt" 
// ) 

type Value interface{}

// type Value interface { 
// string() string 
//  
// // Int performs a 'int' type conversion on Value, and returns 
// // the int. If it isn't a int value, it will panic. 
// // It is a convenience method for 'int(SomeValue.(gribble.Int))'. 
// Int() int 
//  
// // Float performs a 'float64' type conversion on Value, and returns 
// // the float64. If it isn't a float64 value, it will panic. 
// // It is a convenience method for 'float64(SomeValue.(gribble.Float))'. 
// Float() float64 
//  
// // String performs a 'string' type conversion on Value, and returns 
// // the string. If it isn't a string value, it will panic. 
// // It is a convenience method for 'string(SomeValue.(gribble.String))'. 
// String() string 
// } 
//  
// type String string 
//  
// func (v String) string() string { 
// return string(v) 
// } 
//  
// func (v String) Int() int { 
// panic("Cannot call 'Int' on string value.") 
// } 
//  
// func (v String) Float() float64 { 
// panic("Cannot call 'Float' on string value.") 
// } 
//  
// func (v String) String() string { 
// return string(v) 
// } 
//  
// type Int int 
//  
// func (v Int) string() string { 
// return fmt.Sprintf("%d", int(v)) 
// } 
//  
// func (v Int) Int() int { 
// return int(v) 
// } 
//  
// func (v Int) Float() float64 { 
// panic("Cannot call 'Float' on int value.") 
// } 
//  
// func (v Int) String() string { 
// panic("Cannot call 'String' on int value.") 
// } 
//  
// type Float float64 
//  
// func (v Float) string() string { 
// return fmt.Sprintf("%0.2f", float64(v)) 
// } 
//  
// func (v Float) Int() int { 
// panic("Cannot call 'Int' on float value.") 
// } 
//  
// func (v Float) Float() float64 { 
// return float64(v) 
// } 
//  
// func (v Float) String() string { 
// panic("Cannot call 'String' on float value.") 
// } 
