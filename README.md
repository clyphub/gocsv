Go CSV
=====

The GoCSV package aims to provide easy serialization and unserialization functions to use CSV in Go from ```os.File, string or []byte```

API and techniques inspired from http://labix.org/mgo

Full example
=====

Consider the following CSV file

```csv

client_id,client_name,client_age
1,Jose,42
2,Daniel,26
3,Vincent,32

```

Easy binding in Go!
---

```go

package main

import (
	"fmt"
	"gocsv"
	"os"
)

type Client struct { // Our example struct, you can use "-" to ignore a field
	Id      string `csv:"id"`
	Name    string `csv:"name"`
	Age     string `csv:"age"`
	NotUsed string `csv:"-"`
}

func main() {
	file, error := os.OpenFile("clients.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if error != nil {
		panic(error)
	}
	defer file.Close()
	clients := []Client{}

	if err := gocsv.UnmarshalFile(file, &clients); err != nil { // Load the first clients from file
		panic(err)
	}

	for _, client := range clients {
		fmt.Println("Hello", client.Name)
	}

	if _, err := file.Seek(0, 0); err != nil { // Go to the start of the file
		panic(err)
	}

	clients = append(clients, Client{Id: "12", Name: "John", Age: "21"}) // Add clients
	clients = append(clients, Client{Id: "13", Name: "Fred"})
	clients = append(clients, Client{Id: "14", Name: "James", Age: "32"})
	clients = append(clients, Client{Id: "15", Name: "Danny"})

	if err := gocsv.MarshalFile(&clients, file); err != nil { // Save all clients back to the CSV file
		panic(err)
	}

}

```

Customizable CSV Reader / Writer
---

```go

func main() {
	...
	
	gocsv.SetCSVReader(func(in io.Reader) *csv.Reader {
    	//return csv.NewReader(in)
    	return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
    })

    ...

    gocsv.UnmarshalFile(file, &clients)

    ...

    gocsv.SetCSVWriter(func(out io.Writer) *csv.Writer {
    	return csv.NewWriter(out)
    })

    ...

    gocsv.MarshalFile(&clients, file)

	...
}

```