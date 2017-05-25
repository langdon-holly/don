'use strict'

// Dependencies

; var
    ps = require('list-parsing')
    , parser = require('./don-parse.js')
    , _ = require('lodash')
    , nat = require('./nat.js')

// Stuff

; var exports = module.exports;

function apply(fn, arg) {
  function apply(fn, arg) {
    return (
      fn[0] === fnLabel
      ? fn[1](arg)
      : fn[0] === listLabel
        ? arg[0] !== intLabel
          ? Null("Argument to list must be integer")
          : arg[1] < 0 || arg[1] >= fn[1].length
            ? Null("Array index out of bounds")
            : fn[1][arg[1]]
        : Null("Tried to apply a non-function"))}
  return _.reduce(arguments, apply);}
exports.apply = apply;

function fnOfType(type, fn) {
  return (
    makeFn
    ( function(arg)
      { if (arg[0] !== type) return Null("typed function received garbage");
        return fn(arg[1])}))}

function makeFn(fn) {
  return [fnLabel, fn]}

function constFn(val) {return makeFn(_.constant(val))}

function makeList() {
  return [listLabel, Array.from(arguments)]}

function just(val) {
  return makeList(True, val)}

function isTrue(val) {
  var trueVal = [symLabel, {}];
  return apply(val, makeList(trueVal, [symLabel, {}])) === trueVal}

function isString(val) {
  return (
    val[0] === listLabel
    && _.every(val[1], function(elem) {return elem[0] === charLabel}))}

function strVal(list) {
  if (list[0] !== listLabel) return Null();
  return (
    list[1].reduce
    ( function(soFar, chr) {return soFar + charToStr(chr)}
    , ''))}

function stringIs(list, str) {
  return strVal(list) === str}

function charToStr(Char)
{ if (Char[0] !== charLabel)
    return Null("charToStr nonchar: " + strVal(toString(Char)))
; return String.fromCodePoint(Char[1])}

function strToChars(str)
{ return (
    [ listLabel
    , Array.from(str).map
      (function(chr) {return [charLabel, chr.codePointAt(0)]})])}

function eq(val0, val1)
{ return (
    val0[0] === val1[0]
    &&
      ( val0[1] === val1[1]
        ||
          val0[0] === listLabel
          && val0[1].length == val1[1].length
          && _.every
             ( val0[1]
             , function(elem, index) {return eq(elem, val1[1][index])})
        ||
          isString(val0) && isString(val1) && strVal(val0) === strVal(val1)
        || val0[0] === ASTPrecomputedLabel && eq(val0[1], val1[1])
        ||
          val0[0] === callLabel
          && eq(val0[1][0], val1[1][0])
          && eq(val0[1][1], val1[1][1])))}

function parseTreeToAST(pt) {
  var label = pt[0];
  var data = pt[1];

  if (label == 'char') return [charLabel, pt[1]];
  if (label == 'call')
    return (
      _.reduce
      ( data
      , function(applied, arg)
          {return [callLabel, [applied, parseTreeToAST(arg)]]}
      , [ASTPrecomputedLabel, makeFn(function(arg){return arg})]));
  if (label == 'list')
    return (
      [ callLabel
      , [ listVar
        , [ ASTPrecomputedLabel
          , [ listLabel
            , data
              .map(parseTreeToAST)]]]]);
  if (label == 'braced')
    return (
      [ callLabel
      , [ bracedVar
        , [ ASTPrecomputedLabel
          , [listLabel, data.map(parseTreeToAST)]]]]);
  if (label === 'heredoc')
    return [ASTPrecomputedLabel, [listLabel, data.map(parseTreeToAST)]];

  if (label === 'quote') return [ASTPrecomputedLabel, parseTreeToAST(data)];

  if (label === 'quoted-list') return [listLabel, data.map(parseTreeToAST)];

  return Null('unknown parse-tree type "' + label)}

function parseStr(str) {
  var parsed = parser(str);
  if (!parsed[0]) return parsed;

  return [true, parseTreeToAST(parsed[1]), parsed[2]]}
exports.parse = parseStr;

function ttyLog() {
  if (process.stdout.isTTY) console.log.apply(this, arguments)}

var fnLabel = {};
exports.fnLabel = fnLabel;

var listLabel = {};
exports.listLabel = listLabel;

var intLabel = {};
exports.intLabel = intLabel;

var charLabel = {};
exports.charLabel = charLabel;

var symLabel = {};
exports.symLabel = symLabel;

var ASTPrecomputedLabel = {};
exports.ASTPrecomputedLabel = ASTPrecomputedLabel;

var unitLabel = {};
exports.unitLabel = unitLabel;
var unit = [unitLabel];
exports.unit = unit;

var callLabel = {};
exports.callLabel = callLabel;

var listVar = [symLabel, {}];
exports.listVar = listVar;

var bracedVar = [symLabel, {}];
exports.bracedVar = bracedVar;

var Null = function() {
  ttyLog.apply(this, arguments);
  throw new Error("diverging…");
  while (true) {}};
exports.Null = Null;

var False
= makeFn
  ( function(consequent)
      {return makeFn(function(alternative) {return alternative})});

var True
= makeFn
  ( function(consequent)
      {return makeFn(function(alternative) {return consequent})});

var nothing = makeList(False);

var Eval
= makeFn
  ( function(env)
      { return (
          makeFn
          ( function(expr)
              { if (expr[0] === callLabel)
                  return (
                    apply(Eval, env, expr[1][0], apply(Eval, env, expr[1][1])))

                ; if (expr[0] === ASTPrecomputedLabel) return expr[1]

                ; if (expr[0] === symLabel || isString(expr))
                  return apply(env, expr, env)

                ; return Null(expr)}))})
; exports.Eval = Eval

; var topEval = function(ast) {
  //var calls = []
  //; while (continuing.length > 0)
  ; return apply(Eval, initEnv, ast)}
exports.topEval = topEval;

function toString(arg)
{ if (arg[0] === charLabel)
    return (
      [ listLabel
      , [ strToChars("'")[1][0]
        , arg
        , strToChars("'")[1][0]]]);

  if (arg[0] === intLabel)
    return (
      [ listLabel
      , _.toArray(arg[1].toString()).map
        ( function(chr)
          {return strToChars(chr)[1][0]})
        .concat(strToChars(' ')[1])]);

  if (arg[0] === listLabel)
    return (
      [ listLabel
      , strToChars('[')[1].concat
        ( _.reduce
          ( arg[1].map
            (function (o) {return toString(o)[1]})
          , function(soFar, elem, idx)
            {return (
               idx === 0
               ? elem
               : soFar.concat
                 (elem))}
          , [])
        , strToChars(']')[1])]);

  if (arg[0] === ASTPrecomputedLabel)
    return (
      [ listLabel
      , [[charLabel, 34]]
        .concat(toString(arg[1])[1])]);

  if (arg[0] === unitLabel)
    return (
      [ listLabel
      , [117, 110, 105, 116, 32]
        .map(function(chr) {return [charLabel, chr]})])

  if (arg[0] === callLabel)
    return (
      [ listLabel
      , [[charLabel, 92]]
        .concat(toString(arg[1][0])[1])
        .concat(toString(arg[1][1])[1])]);

  if (arg[0] === fnLabel)
    return strToChars("(fn ... )");

  if (arg[0] === symLabel)
    return strToChars("(sym ... )");

  return Null("->str unknown type")}

var initEnv
= makeFn
  ( function(Var)
    { if (Var[0] === symLabel) {
        if (Var === listVar)
          return (
            makeFn
            ( function(env)
              { return (
                  fnOfType
                  ( listLabel
                  , function(arg)
                      { return (
                          [ listLabel
                          , arg.map
                            ( function(elem)
                              { return apply(Eval, env, elem)})])}))}));

        if (Var === bracedVar)
          return (
            constFn
            ( fnOfType
              ( listLabel
              , function(arg)
                  { var argsO
                    = apply(initEnv, listVar, initEnv, [listLabel, arg])
                  ; if (argsO[0] !== listLabel)
                      return Null("Listing returned nonlist");
                  ; var args = argsO[1];
                  ; if (args.length % 2 != 0)
                      return Null("Tried to brace oddity")
                  ; var pairs = _.chunk(args, 2)
                  ; return (
                    makeFn
                    (function (arg)
                       { var toReturn = nothing
                       ; _.forEach
                         ( pairs
                         , function(pair)
                             { if
                                 ( arg === pair[0]
                                   || isString(arg)
                                      && isString(pair[0])
                                      && strVal(arg) === strVal(pair[0]))
                                 toReturn = just(pair[1])})
                       ; return toReturn}))})));

        return Null("symbol variable not found in environment")}

      if (isString(Var)) {
        var thisIsDumb = function () {

//          function default0(pt) {
//            if (pt[0]) return pt[1];
//              return 0}
//    
//          function default1(pt) {
//            if (pt[0]) return pt[1];
//              return 1}
//    
//          function multParts(pt) {
//            return pt[0] * pt[1]}
//    
//          function addParts(pt) {
//            return pt[0] + pt[1]}
//    
//          function digitToNum(chr) {
//            if (chr === '0') return 0;
//            if (chr === '1') return 1;
//            if (chr === '2') return 2;
//            if (chr === '3') return 3;
//            if (chr === '4') return 4;
//            if (chr === '5') return 5;
//            if (chr === '6') return 6;
//            if (chr === '7') return 7;
//            if (chr === '8') return 8;
//            if (chr === '9') return 9;
//            if (chr === 'A' || chr === 'a') return 10;
//            if (chr === 'B' || chr === 'b') return 11;
//            if (chr === 'C' || chr === 'c') return 12;
//            if (chr === 'D' || chr === 'd') return 13;
//            if (chr === 'E' || chr === 'e') return 14;
//            if (chr === 'F' || chr === 'f') return 15}
//    
//          function digitsToNum(base) {
//            return function(digits) {
//                     if (digits.length == 0) return 0;
//                     if (digits.length == 1)
//                       return digitToNum(digits[0]);
//                     return digitsToNum(base)(digits.slice(
//                                                0,
//                                                digits.length - 1))
//                            * base
//                            + digitToNum(digits[digits.length - 1])}}
//    
//          function fracDigitsToNum(base) {
//            return function(digits) {
//              return digitsToNum(1 / base)(digits.reverse()) / base}}
//    
//          function digit(base) {
//            if (base == 2) return ps.or(ps.string('0'),
//                                        ps.string('1'));
//            if (base == 8) return ps.or(digit(2),
//                                        ps.string('2'),
//                                        ps.string('3'),
//                                        ps.string('4'),
//                                        ps.string('5'),
//                                        ps.string('6'),
//                                        ps.string('7'));
//            if (base == 10) return ps.or(digit(8),
//                                         ps.string('8'),
//                                         ps.string('9'));
//            if (base == 16) return ps.or(digit(10),
//                                         ps.string('A'),
//                                         ps.string('a'),
//                                         ps.string('B'),
//                                         ps.string('b'),
//                                         ps.string('C'),
//                                         ps.string('c'),
//                                         ps.string('D'),
//                                         ps.string('d'),
//                                         ps.string('E'),
//                                         ps.string('e'),
//                                         ps.string('F'),
//                                         ps.string('f'))}
//    
//          function digits(base) {
//            return ps.many1(digit(base))}
//    
//          var signParser = ps.mapParser(ps.or(ps.string('+'),
//                                              ps.string('-')),
//                                        function (pt) {
//                                          return pt === '+' ? 1
//                                                            : -1});
//    
//          var numPartParserBase = function(base) {
//            return ps.or(
//              ps.mapParser(
//                digits(base),
//                digitsToNum(base)),
//              ps.mapParser(
//                ps.seq(
//                  ps.mapParser(
//                    ps.opt(
//                      ps.mapParser(
//                        digits(base),
//                        digitsToNum(base))),
//                    default0),
//                  ps.before(
//                    ps.string('.'),
//                    ps.mapParser(
//                      digits(base),
//                      fracDigitsToNum(base)))),
//                addParts))};
//    
//          var urealParserBase = function(base) {
//            var prefix = base == 2 ? ps.string('0b') :
//                         base == 8 ? ps.string('0o') :
//                         base == 16 ? ps.string('0x') :
//                                      ps.or(ps.nothing,
//                                            ps.string('0d'));
//    
//            return ps.before(prefix,
//                             ps.mapParser(
//                               ps.seq(
//                                 numPartParserBase(base),
//                                 ps.mapParser(
//                                   ps.opt(
//                                     ps.mapParser(
//                                       ps.before(
//                                         ps.or(ps.string('e'),
//                                               ps.string('E')),
//                                         ps.mapParser(
//                                           ps.seq(
//                                             ps.mapParser(
//                                               ps.opt(signParser),
//                                               default1),
//                                             numPartParserBase(base)),
//                                           multParts)),
//                                       function (pt) {
//                                         return Math.pow(base, pt)})),
//                                   default1)),
//                               multParts))}
//    
//          var urealParser = ps.or(urealParserBase(2),
//                                  urealParserBase(8),
//                                  urealParserBase(10),
//                                  urealParserBase(16));
//    
//          var realParser = ps.mapParser(ps.seq(ps.mapParser(
//                                                 ps.opt(signParser),
//                                                 default1),
//                                               urealParser),
//                                        multParts);
//    
//          var numParser
//            = ps.or(ps.mapParser(realParser,
//                                 function (pt) {
//                                   return [pt, 0]}),
//                    ps.after(ps.seq(realParser,
//                                    ps.mapParser(
//                                      ps.seq(
//                                        signParser,
//                                        ps.mapParser(
//                                          ps.opt(urealParser),
//                                          default1)),
//                                      multParts)),
//                             ps.string('i')),
//                    ps.mapParser(ps.after(ps.mapParser(
//                                            ps.seq(
//                                              ps.mapParser(
//                                                ps.opt(signParser),
//                                                default1),
//                                              ps.mapParser(
//                                                ps.opt(urealParser),
//                                                default1)),
//                                            multParts),
//                                          ps.string('i')),
//                                 function (pt) {
//                                   return [0, pt]}));

//          if (maybeStr[1].charAt(0) === '"')
//            return valObj(strLabel,
//                          maybeStr[1].slice(1, maybeStr[1].length));

//          var varParts = maybeStr[1].split(':');
//          if (varParts.length >= 2) {
//            return _.reduce(varParts.slice(1, varParts.length),
//                            function(fn, argument) {
//                              return apply(fn,
//                                           valObj(strLabel,
//                                                  argument))},
//                            apply(env,
//                                  valObj(strLabel, varParts[0])))}

          //if (Var[1][0][1] === '"'.codePointAt(0))
          //  return [listLabel, Var[1].slice(1, Var[1].length)];

          if (stringIs(Var, 'fn'))
            return (
              makeFn
              ( function(env)
                { return (
                    makeFn
                    ( function(param)
                      { var newEnv
                        = makeFn
                          ( function(dynEnv)
                            { return (
                                makeFn
                                ( function(Var)
                                  { return (
                                      eq(Var, param)
                                      ? arg
                                      : apply(env, Var, dynEnv))}))})
                      ; return (
                          makeFn
                          ( function(body)
                            { return (
                                makeFn
                                ( function(arg)
                                  { return (
                                      apply(Eval, newEnv, body))}))}))}))}));

          if (stringIs(Var, '+'))
            return (
              constFn
              ( fnOfType
                ( listLabel
                , function (args)
                  { return (
                      _.reduce
                      ( args
                      , function (arg0, arg1
                        ){
                          if (arg1[0] !== intLabel) return Null();
                          return [intLabel, arg0[1] + arg1[1]]}
                      , [intLabel, 0]))})));

          if (stringIs(Var, '-'))
            return (
              constFn
              ( fnOfType
                ( listLabel
                , function (args)
                  { if (args.length === 0) return [intLabel, -1];

                    if (args[0][0] !== intLabel) return Null();
                    if (args.length === 1)
                      return [intLabel, -args[0][1]];

                    return (
                      _.reduce
                      ( args
                      , function (arg0, arg1
                        ){
                          if (arg1[0] !== intLabel)
                            return Null();
                          return [intLabel, arg0[1] - arg1[1]]}))})));

          if (stringIs(Var, '<'))
            return (
              constFn
              ( fnOfType
                ( intLabel
                , function(arg0)
                    { return (
                        fnOfType
                        ( intLabel
                        , function(arg1)
                            {return arg0 < arg1 ? True : False}))})));

          if (stringIs(Var, '='))
            return (
              constFn
              ( makeFn
                ( function(arg0)
                  { return (
                      makeFn
                      ( function(arg1)
                        {return eq(arg0, arg1) ? True : False}))})));

          if (stringIs(Var, "env")) return makeFn(function(env) {return env});

          if (stringIs(Var, "eval")) return constFn(Eval);

          if (stringIs(Var, "print"))
            return (
              constFn
              ( makeFn(function(arg) {
                  return (
                    isString(arg)
                    ? process.stdout.write(strVal(arg))
                    : process.stdout.write(strVal(toString(arg)))
                    , unit)})));

          if (stringIs(Var, "->str"))
            return constFn(makeFn(toString));

          if (stringIs(Var, "str->unicode"))
            return (
              constFn
              ( makeFn
                ( function(arg)
                  { if (!isString(arg)) return Null();
                    return (
                      [ listLabel
                      , arg[1].map
                        ( function(Char)
                          {return [intLabel, Char[1]]})])})));

          if (stringIs(Var, "unicode->str"))
            return (
              constFn
              ( makeFn
                ( function(arg)
                  { if (arg[0] !== listLabel) return Null();

                    return (
                      [ listLabel
                      , arg[1].map
                        ( function (codepoint)
                          { if (codepoint[0] !== intLabel) return Null()
                          ; return [charLabel, codepoint[1]]})])})));

          if (stringIs(Var, "length"))
            return (
              constFn
              ( fnOfType
                ( listLabel
                , function(arg) {return [intLabel, arg.length]})));

          if (stringIs(Var, "->list"))
            return (
              constFn
              ( fnOfType
                ( fnLabel
                , function(fn)
                    {return (
                       fnOfType
                       ( intLabel
                       , function(length)
                           { if (length < 0) return Null();

                             var toReturn = [];
                             for (var i = 0; i < length; i++)
                               {toReturn.push(fn([intLabel, i]))}
                             return [listLabel, toReturn]}))})));

          if (stringIs(Var, "true")) return constFn(True);

          if (stringIs(Var, "false")) return constFn(False);

          if (stringIs(Var, "unit")) return constFn(unit);

          var varStr = strVal(Var);
          if (/^(\-|\+)?[0-9]+$/.test(varStr))
            return constFn([intLabel, parseInt(varStr, 10)]);

          return (
            Null
            ( 'string variable not found in environment: "'
              + strVal(Var)))};
        return thisIsDumb()}

      return Null("unknown variable: " + strVal(toString(Var)))});
exports.initEnv = initEnv;

Error.stackTraceLimit = Infinity;

