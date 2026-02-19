package main

import "os"

func main() {
    f, err := os.Open("test.txt")
    if err != nil {
    	fmt.Println("I could not open that text file");
    }

}
