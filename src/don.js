#!/usr/bin/env node

'use strict';

var don = require('./don-core.js');
var program = require('commander');
var fs = require('fs');

program.parse(process.argv);

function indexToLineColumn(index, string) {
  var arr = Array.from(string), line = 1, col = 1;
  for (var i = 0; i < arr.length; i++) {
    if (i == index) return [line, col];
    if (string[i] === '\n') {
      line++;
      col = 1;}
    else col++;}
  throw new RangeError
            ("indexToLineColumn: index=" + index + " is out of bounds");}

fs.readFile
( program.args.length > 0 ? program.args[0] : 0
, 'utf8'
, function(err, data)
    { if (err) {
        console.log("Couldn't read file");
        process.exit(1);}

      var parsed = don.parse(data);

      if (parsed[0]) {
        don.topEval.apply(this, parsed.slice(1));
        process.exit(0);}
      else {
        var errAt = parsed[1];
        if (errAt == -1)
          console.log
          ( "Syntax error: "
           + program.args[0]
           + " should be at least "
           + (Array.from(data).length + 1)
           + " characters");
        else if (errAt == 0) console.log("Error in the syntax");
        else {
          var lineCol = indexToLineColumn(parsed[1] - 1, data);
          console.log("Syntax error at "
                      + program.args[0]
                      + " "
                      + lineCol[0]
                      + ","
                      + lineCol[1]
                      + ":");
          console.log(data.split('\n')[lineCol[0] - 1]
                      + "\n"
                      + " ".repeat(lineCol[1] - 1)
                      + "^");}
        process.exit(2);}});

