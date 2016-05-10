#!/usr/bin/env node

'use strict';

var parser = require('./don-parse.js');
var don = require('./don-core.js');
var program = require('commander');
var fs = require('fs');

program.parse(process.argv);

fs.readFile(program.args[0], 'utf8', function(err, data) {
  if (err) {
    console.log("Couldn't read file");
    process.exit(1);}

  var parsed = don.parse(data);

  if (parsed[0]) {
    don.topEval(parsed[1]);
    process.exit(0);}
  else {
    console.log("Syntax error");
    process.exit(2);}});

