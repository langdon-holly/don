'use strict';

// Dependencies

var _ = require('lodash');

// Polyfill

// from https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/from
// Production steps of ECMA-262, Edition 6, 22.1.2.1
// Reference: https://people.mozilla.org/~jorendorff/es6-draft.html#sec-array.from
if (!Array.from) {
  Array.from = (function () {
    var toStr = Object.prototype.toString;
    var isCallable = function (fn) {
      return typeof fn === 'function' || toStr.call(fn) === '[object Function]';
    };
    var toInteger = function (value) {
      var number = Number(value);
      if (isNaN(number)) { return 0; }
      if (number === 0 || !isFinite(number)) { return number; }
      return (number > 0 ? 1 : -1) * Math.floor(Math.abs(number));
    };
    var maxSafeInteger = Math.pow(2, 53) - 1;
    var toLength = function (value) {
      var len = toInteger(value);
      return Math.min(Math.max(len, 0), maxSafeInteger);
    };

    // The length property of the from method is 1.
    return function from(arrayLike/*, mapFn, thisArg */) {
      // 1. Let C be the this value.
      var C = this;

      // 2. Let items be ToObject(arrayLike).
      var items = Object(arrayLike);

      // 3. ReturnIfAbrupt(items).
      if (arrayLike == null) {
        throw new TypeError("Array.from requires an array-like object - not null or undefined");
      }

      // 4. If mapfn is undefined, then let mapping be false.
      var mapFn = arguments.length > 1 ? arguments[1] : void undefined;
      var T;
      if (typeof mapFn !== 'undefined') {
        // 5. else
        // 5. a If IsCallable(mapfn) is false, throw a TypeError exception.
        if (!isCallable(mapFn)) {
          throw new TypeError('Array.from: when provided, the second argument must be a function');
        }

        // 5. b. If thisArg was supplied, let T be thisArg; else let T be undefined.
        if (arguments.length > 2) {
          T = arguments[2];
        }
      }

      // 10. Let lenValue be Get(items, "length").
      // 11. Let len be ToLength(lenValue).
      var len = toLength(items.length);

      // 13. If IsConstructor(C) is true, then
      // 13. a. Let A be the result of calling the [[Construct]] internal method of C with an argument list containing the single item len.
      // 14. a. Else, Let A be ArrayCreate(len).
      var A = isCallable(C) ? Object(new C(len)) : new Array(len);

      // 16. Let k be 0.
      var k = 0;
      // 17. Repeat, while k < len… (also steps a - h)
      var kValue;
      while (k < len) {
        kValue = items[k];
        if (mapFn) {
          A[k] = typeof T === 'undefined' ? mapFn(kValue, k) : mapFn.call(T, kValue, k);
        } else {
          A[k] = kValue;
        }
        k += 1;
      }
      // 18. Let putStatus be Put(A, "length", len, true).
      A.length = len;
      // 20. Return A.
      return A;
    };
  }());
}

// Stuff

var exports = module.exports;

function seq() {
  var args = arguments;

  if (args.length == 0) return mapParser(nothing,
                                         function(pt) {
                                           return [];});
  if (args.length == 1) return mapParser(args[0],
                                         function(pt) {
                                           return [pt];});
  if (args.length == 2) {
    var contFirst = doomed(args[0]) || doomed(args[1])
                    ? fail
                    : {parseChar:
                         function (chr) {
                           return seq(args[0].parseChar(chr), args[1]);},
                       result: [false],
                       noMore: args[0].noMore && args[1].noMore,
                       futureSuccess: args[0].futureSuccess
                                      && (args[1].result[0]
                                          || alwaysSuccessful(args[1]))};
    return args[0].result[0] ? or(contFirst,
                                  mapParser(args[1],
                                            function (pt) {
                                              return [args[0].result[1], pt];}))
                             : contFirst;}
  return mapParser(seq(args[0],
                       seq.apply(this,
                                 Array.from(args).slice(1, args.length))),
                   function(arr) {return [arr[0]].concat(arr[1]);});}
exports.seq = seq;

function character(chr0) {
  return {parseChar: function (chr1) {
            return chr0 === chr1 ? {parseChar: function () {return fail;},
                                    result: [true, chr0],
                                    noMore: true,
                                    futureSuccess: false}
                                 : fail},
          result: [false],
          noMore: false,
          futureSuccess: false};}

function ciCharacter(chr0) {
  chr0 = chr0.toUpperCase();
  return {parseChar: function (chr1) {
            var chr2 = chr1.toUpperCase();
            return chr0 === chr2 ? {parseChar: function () {return fail;},
                                    result: [true, chr1],
                                    noMore: true,
                                    futureSuccess: false}
                                 : fail},
          result: [false],
          noMore: false,
          futureSuccess: false};}

function string(str) {
  return mapParser(seq.apply(this, str.split('').map(function (chr) {
                     return character(chr);})),
                   function concat(arr) {
                     if (arr.length == 0) return '';
                     return arr[0]
                            + concat(arr.slice(1, arr.length));});}
exports.string = string;

function ciString(str) {
  return mapParser(seq.apply(this, str.split('').map(function (chr) {
                     return ciCharacter(chr);})),
                   function concat(arr) {
                     if (arr.length == 0) return '';
                     return arr[0]
                            + concat(arr.slice(1, arr.length));});}
exports.ciString = ciString;

var fail = {parseChar: function() {return fail;},
            result: [false],
            noMore: true,
            futureSuccess: false};
exports.fail = fail;

var anything = {parseChar: function(chr) {
                  return mapParser(anything,
                                   function (pt) {return chr + pt;});},
                result: [true, ''],
                noMore: false,
                futureSuccess: true};
exports.anything = anything

function many(parser) {
  return {parseChar: function(chr) {
            return many1(parser).parseChar(chr);},
          result: [true, []],
          noMore: parser.noMore,
          futureSuccess: parser.futureSuccess};}
exports.many = many;

function manyCount(fn, start) {
  start = start || 0;

  return {parseChar: function(chr) {
            return many1Count(fn, start).parseChar(chr);},
          result: [true, []],
          noMore: fn(start).noMore,
          futureSuccess: fn(start).futureSuccess};}
exports.manyCount = manyCount;

function many1(parser) {
  return mapParser(seq(and(not(nothing), parser), many(parser)),
                   function(pt) {return [pt[0][1]].concat(pt[1]);});}
exports.many1 = many1;

function many1Count(fn, start) {
  start = start || 0;

  return mapParser(seq(and(not(nothing), fn(start)), manyCount(fn, start + 1)),
                   function(pt) {return [pt[0][1]].concat(pt[1]);});}
exports.many1Count = many1Count;

function parse(parser, str) {
  if (doomed(parser)) return [false];
  if (str.length == 0) return parser.result;
  return parse(parser.parseChar(str.charAt(0)), str.slice(1, str.length));}
exports.parse = parse;

function longestMatch(parser, str) {
  if (doomed(parser)) return [[false]];

  var toReturn = [[false]];

  if (parser.result[0]) toReturn = [parser.result, 0];
  for (var index = 0; index < str.length && !doomed(parser); index++) {
    //console.log("parsing {" + str.charAt(index) + "}");
    parser = parser.parseChar(str.charAt(index));
    if (parser.result[0]) toReturn = [parser.result, index + 1];
    if (!toReturn[0][0]) toReturn[1] = index;}
  return toReturn;}
exports.longestMatch = longestMatch;

function shortestMatch(parser, str) {
  if (doomed(parser)) return [[false]];
  if (parser.result[0]) return [parser.result, 0];

  for (var index = 0; index < str.length && !doomed(parser); index++) {
    parser = parser.parseChar(str.charAt(index));
    if (parser.result[0]) return [parser.result, index + 1];}
  return [[false], index];}
exports.shortestMatch = shortestMatch;

function sepByCount(elemFn, sepFn, atLeast1) {
  if (!atLeast1) return or(mapParser(nothing,
                                     function(pt) {
                                       return [];}),
                           sepByCount(elemFn, sepFn, true));

  return mapParser(seq(elemFn(0),
                       manyCount(function(index) {
                                   return before(sepFn(index), elemFn(index));},
                                 1)),
                   function(pt) {
                     return [pt[0]].concat(pt[1])});}
exports.sepByCount = sepByCount;

function sepBy(element, separator) {
  return sepByCount(function() {return element;},
                    function() {return separator;});}
exports.sepBy = sepBy;

function sepBy1(element, separator) {
  return sepByCount(function() {return element;},
                    function() {return separator;},
                    true);}
exports.sepBy1 = sepBy1;

function opt(parser) {
  return or(
    mapParser(nothing,
              function() {
                return [false];}),
    mapParser(parser,
              function(pt) {
                return [true, pt];}));}
exports.opt = opt;

function mapParser(parser, fn) {
  if (doomed(parser)) return fail;
  return {parseChar: function(chr) {
            return mapParser(parser.parseChar(chr), fn);},
          result: parser.result[0] ? [true, fn(parser.result[1])]
                                   : [false],
          noMore: parser.noMore,
          futureSuccess: parser.futureSuccess};}
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

function between(parser0, parser1) {
  if (arguments.length == 0) return nothing;
  if (arguments.length == 1)
    return mapParser(parser0, function(pt) {return [];});
  if (arguments.length == 2)
    return mapParser(before(parser0, parser1), function(pt) {return [pt];});
  return mapParser(seq(between(parser0, parser1),
                       between.apply(this,
                                     Array.prototype.slice.call(arguments, 2))),
                   _.flatten);}
exports.between = between;

function or() {
  var args = _.filter(arguments, function (parser) {return !doomed(parser);}),
      parser0 = args[0],
      parser1 = args[1];

  if (args.length == 0) return fail;
  if (args.length == 1) return parser0;
  if (args.length == 2)
    return {parseChar: function(chr) {
              return or(parser0.parseChar(chr),
                        parser1.parseChar(chr));},
            result: parser0.result[0] ? parser0.result
                                      : parser1.result,
            noMore: parser0.noMore && parser1.noMore,
            futureSuccess: parser0.futureSuccess || parser1.futureSuccess};
  return or(args[0],
            or.apply(this,
                     Array.from(args).slice(1, args.length)));}
exports.or = or;

function and(parser0, parser1) {
  var args = arguments;

  if (args.length == 0) return mapParser(anything,
                                         function() {return [];});
  if (args.length == 1) return mapParser(parser0,
                                         function(pt) {return [pt];});
  if (args.length == 2) {
    if (doomed(parser0)
        || doomed(parser1)
        || !parser0.result[0] && parser1.noMore
        || !parser1.result[0] && parser0.noMore) return fail;
    return {parseChar: function(chr) {
              return and(parser0.parseChar(chr),
                         parser1.parseChar(chr));},
            result: parser0.result[0] && parser1.result[0]
                    ? [true, [parser0.result[1], parser1.result[1]]]
                    : [false],
            noMore: parser0.noMore || parser1.noMore,
            futureSuccess: parser0.futureSuccess && parser1.futureSuccess};}
  return mapParser(and(parser0,
                       and.apply(this,
                                 Array.from(args).slice(1, args.length))),
                   function(arr) {return [arr[0]].concat(arr[1])});}
exports.and = and;

function strOfLength(len) {
  return len == 0 ? {parseChar: function() {return fail;},
                     result: [true, ''],
                     noMore: true,
                     futureSuccess: false}
                  : {parseChar: function(chr) {
                       return mapParser(strOfLength(len - 1),
                                        function (pt) {
                                          return chr + pt;});},
                     result: [false],
                     noMore: false,
                     futureSuccess: false};}
exports.strOfLength = strOfLength;

function not(parser) {
  return {parseChar: function(chr) {
            return mapParser(not(parser.parseChar(chr)),
                             function(pt) {return chr + pt;});},
          result: parser.result[0] ? [false] : [true],
          noMore: parser.futureSuccess,
          futureSuccess: parser.noMore}}
exports.not = not;

function charNot() {
  var args = arguments;
  return mapParser(and.apply(this, [strOfLength(1)].concat(_.map(args, not))),
                   function(pt) {return pt[0];});}
exports.charNot = charNot;

var nothing = {parseChar: function() {return fail;},
               result: [true, ''],
               noMore: true,
               futureSuccess: false};
exports.nothing = nothing;

var hSpaceChar = or(string('\t'),
                    string(' '),
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
                    string('\u202F'),
                    string('\u205F'),
                    string('\u3000'));
exports.hSpaceChar = hSpaceChar;

var hSpace = many1(hSpaceChar);
exports.hSpace = hSpace;

var vSpaceChar = or(string('\u000A'),
                    string('\u000B'),
                    string('\f'),
                    string('\r'),
                    string('\u0085'),
                    string('\u2028'),
                    string('\u2029'));
exports.vSpaceChar = vSpaceChar;

var vSpace = many1(vSpaceChar);
exports.vSpace = vSpace;

var wsChar = or(hSpaceChar, vSpaceChar);
exports.wsChar = wsChar;

var ws = many1(wsChar);
exports.ws = ws;

function doomed(parser) {
  return !parser.result[0] && parser.noMore;}
exports.doomed = doomed;

function recurseLeft(recursive, nonrecursive, emptyCriteria) {
  return mapParser(recursive.result[0] && emptyCriteria
                   ? seq(opt(nonrecursive), many(recursive))
                   : seq(mapParser(nonrecursive,
                                   function(pt) {return [true, pt];}),
                         many(recursive)),
                   function(pt) {
                     return (pt[0][0]
                             ? [pt[0][1]]
                             : [])
                            .concat(pt[1]);});}
exports.recurseLeft = recurseLeft;

function recurseRight(recursive, nonrecursive, emptyCriteria) {
  return mapParser(recursive.result[0] && emptyCriteria
                   ? seq(many(recursive), opt(nonrecursive))
                   : seq(many(recursive),
                         mapParser(nonrecursive,
                                   function(pt) {return [true, pt];})),
                   function(pt) {
                     return pt[0].concat((pt[1][0]
                                          ? [pt[1][1]]
                                          : []));});}
exports.recurseRight = recurseRight;

function recurse(inTermsOfThis, optimistic) {
  optimistic = optimistic || false;

  function makeThis(matches) {
    if (doomed(matches.parser)) return fail;

    function findCharIndex(chr) {
      var index = matches.characters.length;
      for (var i = 0; i < matches.characters.length; i++) {
        if (matches.characters[i].character === chr) {
          index = i;
          break;}}
      return index;}

    return {parseChar: function(chr) {
                         var index = findCharIndex(chr);
                         var newMatches = matches.characters[index].match;

                         // if chr has not been parsed yet
                         if (index == matches.characters.length) {
                           matches.characters.push(
                             { character: chr,
                               match: { parser: optimistic ? string('') : fail,
                                        characters: []}});
                           newMatches = matches.characters[index].match;
                           newMatches.parsers.parser
                             = matches.parsers
                                      .parser
                                      .parseChar(chr);
                           newMatches.parsers.parsed
                             = true;}

                         // if chr is in the middle of being parsed
                         if (!matches.characters[index].match.parsers.parsed) {
                           newMatches.parsers.defaultParser
                             = matches
                               .parsers
                               .defaultParser
                               .parseChar(chr);
                           newMatches.parsers.undefaultParser
                             = matches
                               .parsers
                               .undefaultParser
                               .parseChar(chr);
                           matches.characters[index].match.parsers.primaryParser
                             = 'default';}

                         return makeThis(matches.characters[index].match);},
            result:
              matches.parsers.primaryParser === 'parser'
              ? matches.parsers.parser.result
              : matches.parsers.parsed
                ? matches.parsers.parser.result[0]
                  === matches.parsers.defaultParser.result[0]
                  ? matches.parsers.parser.result
                  : function() {
                      throw new Error("recursive parser contradicts itself");}()
                : matches.parsers.defaultParser.result,
            noMore: false,
            futureSuccess: false};}

  var defaultEmpty = inTermsOfThis(optimistic ? anything : fail);
  var undefaultEmpty = inTermsOfThis(optimistic ? fail : anything);
  var matches = { parser: defaultEmpty,
                  characters: []};
  var This = makeThis(matches);
  var otherThis = inTermsOfThis(This);
  matches.parsers.parser = otherThis;

  return This;}
exports.recurse = recurse;

