var _ = require('lodash');

var exports = module.exports;

function seq() {
  var args = arguments;

  var toReturn = function(str) {
    function checkSlicing(betweenI) {
      var firstOne = args[0](str.slice(0, betweenI));
      if (firstOne[0]) {
        var lastOne = args[1](str.slice(betweenI, str.length));
        if (lastOne[0]) {
          return [true, [firstOne[1], lastOne[1]]];}}
      return [false];}

    if (args.length == 0) return mapParser(string(''),
                                           function(pt) {
                                             return [];})(str);
    if (args.length == 1) return mapParser(args[0],
                                           function(pt) {
                                             return [pt];})(str);
    if (args.length == 2) {
      if (args[0].len === undefined) {
        for (var betweenI = str.length; betweenI >= 0; betweenI--) {
          var maybe = checkSlicing(betweenI);
          if (maybe[0]) return maybe;}
        return [false];}
      return checkSlicing(args[0].len);}
    return mapParser(seq(args[0],
                         seq.apply(this,
                                   Array.prototype.slice.call(args,
                                                              1,
                                                              args.length))),
                     function(arr) {return [arr[0]].concat(arr[1]);})(str);}
  if (args.length == 2
    && args[0].len !== undefined
    && args[1].len !== undefined) toReturn.len = args[0].len + args[1].len;
  return toReturn;}
exports.seq = seq;

function string(str0) {
  var toReturn = function(str1) {
    if (str0 === str1) return [true, str0];
    return [false];}
  toReturn.len = str0.length;
  return toReturn;}
exports.string = string;

var wsChar = function(str) {
  if (str === '\t' ||
      str === '\u000A' ||
      str === '\u000B' ||
      str === '\f' ||
      str === '\r' ||
      str === ' ' ||
      str === '\u0085' ||
      str === '\u00A0' ||
      str === '\u1680' ||
      str === '\u2000' ||
      str === '\u2001' ||
      str === '\u2002' ||
      str === '\u2003' ||
      str === '\u2004' ||
      str === '\u2005' ||
      str === '\u2006' ||
      str === '\u2007' ||
      str === '\u2008' ||
      str === '\u2009' ||
      str === '\u200A' ||
      str === '\u2028' ||
      str === '\u2029' ||
      str === '\u202F' ||
      str === '\u205F' ||
      str === '\u3000')
    return [true, str]
  return [false];}
wsChar.len = 1;
exports.wsChar = wsChar;

function ws(str) {
  return many1(wsChar)(str);}
exports.ws = ws;

function many(parser) {
  return function(str) {
    if (str === '') return [true, []];
    return many1(parser)(str);}}
exports.many = many;

function many1(parser) {
  return function(str) {
    var firstMatch = longestMatch(parser, str);
    if (!firstMatch[0][0]) return [false];
    if (firstMatch[1] == str.length) return [true, [firstMatch[0][1]]];
    var restMatch = many(parser)(str.slice(firstMatch[1], str.length));
    if (!restMatch[0]) return [false];
    return [true, [firstMatch[0][1]].concat(restMatch[1])];}}
exports.many1 = many1;

function longestMatch(parser, str) {
  for (var end = str.length; end >= 0; end--) {
    var parseResult = parser(str.slice(0, end));
    if (parseResult[0]) return [parseResult, end];}
  return [[false]];}
exports.longestMatch = longestMatch;

function shortestMatch(parser, str) {
  for (var end = 0; end <= str.length; end++) {
    var parseResult = parser(str.slice(0, end));
    if (parseResult[0]) return [parseResult, end];}
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
  return function(str) {
    if (str === '') return [true, [false]];
    return mapParser(parser(str),
                     function(pt) {
                       return [true, pt];});}}
exports.opt = opt;

function mapParser(parser, fn) {
  var toReturn = function(str) {
    var parseResult = parser(str);
    if (!parseResult[0]) return [false];
    return [true, fn(parseResult[1])];}
  toReturn.len = parser.len;
  return toReturn;}
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
  var toReturn = function (str) {
    if (str.length == len) return [true, str];
    return [false];};
  toReturn.len = len;
  return toReturn;}
exports.strOfLength = strOfLength;

function not(parser) {
  return function(str) {
    if (parser(str)[0]) return [false];
    return [true, str];}}
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

