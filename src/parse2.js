var _ = require('lodash');

var exports = module.exports;

function seq() {
  var args = arguments;
}
exports.seq = seq;

function character(chr0) {
  return {parseChar: function (chr1) {
            chr0 === chr1 ? {parseChar: function () {return failParser;},
                             result: [true, chr0]}
                          : failParser},
          result: [false]};}

function string(str) {
  return seq.apply(this, str.slice('').map(function (chr) {
    return character(chr);}));}
exports.string = string;

var failParser = {parseChar: function() {return failParser;},
                  result: [false]};

var wsChar = or(string('\t'),
                string('\u000A'),
                string('\u000B'),
                string('\f'),
                string('\r'),
                string(' '),
                string('\u0085'),
                string('\u00A0'),
                string('\u1680'),
                string('\u2000'),
                string('\u2001'),
                string('\u2002'),
                string('\u2003'),
                string('\u2004'),
                string('\u2005'),
                string('\u2006'),
                string('\u2007'),
                string('\u2008'),
                string('\u2009'),
                string('\u200A'),
                string('\u2028'),
                string('\u2029'),
                string('\u202F'),
                string('\u205F'),
                string('\u3000'));
exports.wsChar = wsChar;

var ws = many1(wsChar);
exports.ws = ws;

function many(parser) {
  return or(string(''),
            many1(parser));}
exports.many = many;

function many1(parser) {
  return function(str) {
    var firstMatch = longestMatch(parser, str);
    if (!firstMatch[0][0]) return [false];
    if (firstMatch[1] == str.length) return [true, [firstMatch[0][1]]];
    var restMatch = parse(many(parser), str.slice(firstMatch[1], str.length));
    if (!restMatch[0]) return [false];
    return [true, [firstMatch[0][1]].concat(restMatch[1])];}}
exports.many1 = many1;

function longestMatch(parser, str) {
  var toReturn = [[false]];

  if (parser.result[0]) toReturn = [parser.result, 0];
  for (var index = 0; index < str.length; index++) {
    parser = parser.parseChar(str.charAt(index));
    if (parser.result[0]) toReturn = [parser.result, index + 1];}
  return toReturn;}
exports.longestMatch = longestMatch;

function shortestMatch(parser, str) {
  if (parser.result[0]) return [parser.result, 0];
  for (var index = 0; index < str.length; index++) {
    parser = parser.parseChar(str.charAt(index));
    if (parser.result[0]) return [parser.result, index + 1];}
  return [[false]];}
exports.shortestMatch = shortestMatch;

function sepBy(element, separator) {
  return or(mapParser(string(''),
                      function(pt) {
                        return [];}),
            sepBy1(element, separator));}
exports.sepBy = sepBy;

function sepBy1(element, separator) {
  return mapParser(seq(element, many(before(separator, element))),
                   function(pt) {
                     return [pt[0]].concat(pt[1])});}
exports.sepBy1 = sepBy1;

function opt(parser) {
  return or(
    mapParser(nothing,
              function() {
                return [false];})
    mapParser(parser,
              function(pt) {
                return [true, pt];}));}
exports.opt = opt;

function mapParser(parser, fn) {
  return {parseChar: function(chr) {
                       return mapParser(parser.parseChar(chr), fn);},
          result: parser.result[0] ? [true, fn(parser.result[1])]
                                   : [false]};}
exports.mapParser = mapParser;

function before(parser0, parser1) {
  return mapParser(seq(parser0, parser1),
                   function(arr) {
                     return arr[1];});}
exports.before = before;

function after(parser0, parser1) {
  return mapParser(seq(parser0, parser1),
                   function(arr) {
                     return arr[0];});}
exports.after = after;

function around(parser0, parser1, parser2) {
  return before(parser0, after(parser1, parser2));}
exports.around = around;

function or(parser0, parser1) {
  var args = arguments;

  var toReturn = function(str) {
    if (args.length == 0) return [false];
    if (args.length == 1) return parser0(str);
    if (args.length == 2) {
      var parseResult = parser0(str);
      if (parseResult[0]) return parseResult;
      return parser1(str);}
    return or(args[0],
              or.apply(this,
                       Array.prototype.slice.call(args,
                                                  1,
                                                  args.length)))(str);}
  if (args.length == 2
      && parser0.len !== undefined
      && parser1.len !== undefined
      && parser0.len == parser1.len) toReturn.len = parser0.len;
  return toReturn;}
exports.or = or;

function and(parser0, parser1) {
  var args = arguments;

  var toReturn = function(str) {
    if (args.length == 0) return [true, []];
    if (args.length == 1) return mapParser(parser0,
                                          function(pt) {return [pt];})(str);
    if (args.length == 2) {
      var firstResult = parser0(str);
      var lastResult = parser1(str);

      if (firstResult[0] && lastResult[1]) return [ true,
                                                    [ firstResult[1],
                                                      lastResult[1]]];
      return [false];}
    return mapParser(and(parser0,
                         and.apply(this,
                                   Array.prototype.slice.call(args,
                                                              1,
                                                              args.length))),
                     function(arr) {return [arr[0]].concat(arr[1])})(str);}
  if (args.length == 2) toReturn.len = parser0.len;
  return toReturn;}
exports.and = and;

function strOfLength(len) {
  return len == 0 ? {parseChar: function() {return failParser;},
                     result: [true, '']}
                  : {parseChar: function(chr) {
                       return mapParser(strOfLength(len - 1),
                                        function (pt) {
                                          return chr + pt;});},
                     result: [false]};}
exports.strOfLength = strOfLength;

function not(parser) {
  return {parseChar: function(chr) {
            return not(parser.parseChar(chr));},
          result: parser.result[0] ? [false] : [true]}}
exports.not = not;

function charNot() {
  return mapParser(and.apply(this, [strOfLength(1)].concat(_.map(arguments,
                                                                 not))),
                   function(arr) {return arr[0]});}
exports.charNot = charNot;

function nothing(str) {
  return string('')(str);}
exports.nothing = nothing;

//console.log(seq(nothing, charNot(wsChar))('k'));

