package main

import (
    "strings"
    "strconv"
    "bufio"
    //"fmt"
    "log"
    "os"
)

func readIndexes(indexesfile string) {
    // open file
    f, err := os.Open(indexesfile)
    if err != nil {
        log.Fatal(err)
    }
    // remember to close the file at the end of the program
    defer f.Close()

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        // do something with a line

        line := scanner.Text()
        eq := strings.IndexRune(line, ':')

        n, _ := strconv.Atoi(line[eq+1:])
        indexes[line[:eq]] = n
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}

func MakeIndexes(mappingsfile string) {
    // open file
    f, err := os.Open(mappingsfile)
    if err != nil {
        log.Fatal(err)
    }
    // remember to close the file at the end of the program
    defer f.Close()

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    var index int = 0
    var target = ""

    for scanner.Scan() {
        // do something with a line
        line := scanner.Text()
        if (line[0] == ' ') {
            // method/field
        } else if (line[1] != ' ') {
            // class
            space := strings.IndexRune(line, ' ')

            //mojmap := line[:space]
            proguard := line[space+4:]

            //fmt.Printf("class: %s = %d\n", proguard, index)

            //TODO use string builder
            target += proguard + strconv.Itoa(index) + "\n"

            //ip := "indexes/" + string(proguard[0]) + ".txt"

            //if os.Exists(ip) {
            //    target := os.ReadFile(ip)
            //} else {
            //    os.WriteFile(ip, "")
            //}
            //TODO  Chunked indexes? Or would that be slower because network/IO?
        }

        index += len(line) + 1
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
    
    os.WriteFile("indexes.txt", []byte(target[:len(target)-1]), 0700)
}
