require('lodash');

function expr(str) {
  if (str.charAt(0) === '(') {
    return form(str);}}

function form(str) {
  return around(seq(string("("),
                    ows),
                sepBy(exprs, ws),
                seq(ows,
                    string(")")));}

function list(str) {
  return around(seq(string("["),
                    ows),
                sepBy(exprs, ws),
                seq(ows,
                    string("]")));}

function seq() {
  return function(str) {
    var args = arguments;

    if (args.length == 0) return mapParser(string(''),
                                           function(pt) {
                                             return [];});
    if (args.length == 1) return mapParser(args[0],
                                           function(pt) {
                                             return [pt];});
    else if (args.length == 2) {
      for (var betweenI = str.length; betweenI >= 0; betweenI--) {
        var firstOne = args[0](str.slice(0, betweenI));
        if (firstOne[0]) {
          var lastOne = args[1](str.slice(betweenI, str.length));
          if (lastOne[0]) {
            return [true, [firstOne[1], lastOne[1]]];}}}
      return [false];}
    else if (args.length > 2) {
      return seq(args[0], seq.apply(this, Array.prototype.slice.call(args, 1, args.length)));}}}

function string(str0) {
  return function(str1) {
    if (str0 === str1) return [true, str0];
    return [false]}}

function ws(str) {
  if (str === ' ' ||
      str === '\n' ||
      str === '\t')
    return [true, str]
  return [false];}

function many(parser) {
  return function(str) {
    if (str === '') return [true, []];
    return many1(str);}}

function many1(parser) {
  return function(str) {
    var firstMatch = longestMatch(parser, str);
    if (!firstMatch[0][0]) return [false];
    if (firstMatch[1] == str.length) return [true, [firstMatch[0][1]]];
    var restMatch = many(parser)(str.slice(firstMatch[1], str.length));
    if (!restMatch[0]) return [false];
    return [true, [firstMatch[0][1]].concat(restMatch[1])];}}

function longestMatch(parser, str) {
  for (var end = str.length; end >= 0; end--) {
    var parseResult = parser(str.slice(0, end));
    if (parseResult[0]) return [parseResult, end];}
  return [[false]];}

function sepBy(element, separator) {
  return mapParser(seq(element, many(before(separator, element))),
                   function(pt) {
                     return [pt[0]].concat(pt[1])});}

function opt(parser) {
  return function(str) {
    if (str === '') return [true, [false]];
    return mapParser(parser(str),
                     function(pt) {
                       return [true, pt];});}}

function mapParser(parser, fn) {
  return function(str) {
    var parseResult = parser(str);
    if (!parseResult[0]) return [false];
    return [true, fn(parseResult[1])];}}

function before(parser0, parser1) {
  return mapParser(seq(parser0, parser1),
                   function(arr) {
                     return arr[1];});}

function after(parser0, parser1) {
  return mapParser(seq(parser0, parser1),
                   function(arr) {
                     return arr[0];});}

function around(parser0, parser1, parser2) {
  return before(parser0, after(parser1, parser2));
}

var ows = opt(ws);

