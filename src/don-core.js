'use strict';

// Dependencies

var ps = require('list-parsing');
var parser = require('./don-parse.js');
var _ = require('lodash');

// Polyfill

if (!Array.prototype.includes) {
  Array.prototype.includes = function(searchElement /*, fromIndex*/ ) {
    'use strict';
    var O = Object(this);
    var len = parseInt(O.length) || 0;
    if (len === 0) {
      return false;
    }
    var n = parseInt(arguments[1]) || 0;
    var k;
    if (n >= 0) {
      k = n;
    } else {
      k = len + n;
      if (k < 0) {k = 0;}
    }
    var currentElement;
    while (k < len) {
      currentElement = O[k];
      if (searchElement === currentElement ||
         (searchElement !== searchElement && currentElement !== currentElement)) { // NaN !== NaN
        return true;
      }
      k++;
    }
    return false;
  };
}

// Stuff

var exports = module.exports;

function mApply(macro, arg, env) {
  if (macro[0] === macroLabel)
    return macro[1](arg, env);

  arg = apply(apply(Eval, arg, env), env, env);

  if (macro[0] === fnLabel)
    return macro[1](arg, env);

  if (macro[0] === listLabel) {
    if (arg[0] !== intLabel)
      return Null("Argument to list must be integer");
    if (arg[1] < 0 || arg[1] >= macro[1].length)
      return Null("Array index out of bounds");
    return macro[1][arg[1]];}

  return Null("Tried to macro-apply a non-macro");}
exports.mApply = mApply;

function apply(fn, arg, env) {
  return fn[0] === macroLabel ? mApply(fn,
                                       [ASTPrecomputedLabel, arg],
                                       env)
         : fn[0] === fnLabel ? fn[1](arg, env)
         : fn[0] === listLabel ? function(){
           if (arg[0] !== intLabel)
             return Null("Argument to list must be integer");
           if (arg[1] < 0 || arg[1] >= fn[1].length)
             return Null("Array index out of bounds");
           return fn[1][arg[1]];}()
         : Null("Tried to apply a non-macro");}
exports.apply = apply;

function fnOfType(type, fn) {
  return makeFn(function(arg, env) {
                  if (arg[0] !== type
                  ) return Null("typed function received garbage");
                  return fn(arg[1], env);})}

function makeFn(fn) {
  return [fnLabel,
          fn];}

function makeList() {
  return [listLabel, Array.from(arguments)];}

function just(val) {
  return makeList(True,
                  val);}

function isTrue(val, env) {
  var trueVal = [symLabel, {}];
  return apply(val, makeList(trueVal, [symLabel, {}]), env) === trueVal;}

function isString(val) {
  return val[0] === listLabel
         && _.every(val[1], function(elem) {return elem[0] === charLabel;});}

function strVal(list) {
  if (list[0] !== listLabel) return Null();
  return list[1].reduce
         ( function(soFar, chr) {return soFar + charToStr(chr);}
         , '');}

function stringIs(list, str) {
  return strVal(list) === str;}

function charToStr(Char)
{ if (Char[0] !== charLabel) return Null()
; return String.fromCodePoint(Char[1])}

function parseTreeToAST(pt) {
  var label = pt[0];

  if (label == 'char') return [charLabel, pt[1]];
  if (label == 'call') return [listLabel, pt[1].map(parseTreeToAST)];
  if (label == 'list') return [ listLabel,
                                [ listVar,
                                  [ listLabel,
                                    pt[1]
                                    .map(parseTreeToAST)]]];
  if (label == 'braceStr')
    return [ listLabel,
             [ preEvalVar,
               [ ASTPrecomputedLabel,
                 [ listLabel,
                   pt
                   [1]
                   .map
                   (function(elem) {
                      return [ASTBraceStrElemLabel,
                              elem[0] === 'expr'
                              ? ['expr', parseTreeToAST(elem[1])]
                              : elem];})]]]];
  if (label === 'heredoc')
    return [ASTPrecomputedLabel, [listLabel, pt[1].map(parseTreeToAST)]];

  return Null('unknown parse-tree type "' + label);}

function parseStr(str) {
  var parsed = parser(str);
  if (!parsed[0][0]) return [false, parsed[0][1]];

  return [true, parseTreeToAST(parsed[0][1])];}
exports.parse = parseStr;

function ttyLog() {
  if (process.stdout.isTTY) console.log.apply(this, arguments);}

var fnLabel = {};
exports.fnLabel = fnLabel;

var macroLabel = {};
exports.macroLabel = macroLabel;

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

var ASTBraceStrElemLabel = {};
exports.ASTBraceStrElemLabel = ASTBraceStrElemLabel;

var unitLabel = {};
exports.unitLabel = unitLabel;
var unit = [unitLabel];
exports.unit = unit;

var preEvalVar = [symLabel, {}];
exports.preEvalVar = preEvalVar;

var braceStrEvalVar = [symLabel, {}];
exports.braceStrEvalVar = braceStrEvalVar;

var listVar = [symLabel, {}];
exports.listVar = listVar;

var Null = function() {
  ttyLog.apply(this, arguments);
  throw new Error("diverging…");
  while (true) {}};
exports.Null = Null;

var False
= [ macroLabel
  , function(arg0, env
    ){ return [fnLabel, function(arg1, env) {return arg1;}];}];

var True
= [ fnLabel
  , function(arg0, env
    ){ return [macroLabel, function(arg1, env) {return arg0;}];}];

var nothing = makeList(False);

var Eval
= makeFn
  ( function(expr
    ){
      return makeFn
             ( function(env
               ){
                 if (expr[0] === listLabel) {
                   if (expr[1][0][0] === charLabel) return apply(env, expr, env);
                   return _
                          .reduce
                          ( expr[1]
                          , function(val, arg
                            ) {return mApply(val, arg, env);}
                          , makeFn(function (arg, env) {return arg;}));}

                 if (expr[0] === ASTPrecomputedLabel) return expr[1];

                 if (expr[0] === symLabel) return apply(env, expr, env);

                 if (expr[0] === charLabel) return Null();

                 return Null();});});
exports.Eval = Eval;

var topEval = function(ast) {
  return apply(apply(Eval, ast, initEnv), initEnv, initEnv);}
exports.topEval = topEval;

var initEnv
  = makeFn(function(Var, env) {
             if (Var[0] === symLabel) {
               if (Var === preEvalVar)
                 return fnOfType
                        ( listLabel
                        , function(arg, env) {
                            return [ listLabel,
                                     _
                                     .reduce
                                     ( _
                                       .map
                                       ( arg,
                                         function(elem) {
                                           if(
                                             elem[0] !== ASTBraceStrElemLabel)
                                             return Null();
                                           return elem[1];}),
                                       function(soFar, elem) {
                                         var toAppend;
                                         if (elem[0] === 'str')
                                           toAppend = elem[1].map(function(chr) {return [charLabel, chr.codePointAt(0)];});
                                         else if (elem[0] === 'expr') {
                                           var
                                             evaled
                                           = apply
                                             ( apply(Eval, elem[1], env)
                                             , env
                                             , env);
                                           if (evaled[0] !== listLabel)
                                             return Null();
                                           toAppend = evaled[1];}
                                         else return Null();
                                         return soFar.concat(toAppend);},
                                       [])];});

               if (Var === listVar)
                 return [macroLabel, function(arg, env) {
                   if (arg[0] !== listLabel) return Null();
                   return [ listLabel
                          , arg[1].map
                            ( function(elem)
                              { return apply
                                       ( apply(Eval, elem, env)
                                       , env
                                       , env);})];}];

               if (Var === braceStrEvalVar) return Eval;

               return Null();}

             if (isString(Var)) {
               var thisIsDumb = function () {

//                 function default0(pt) {
//                   if (pt[0]) return pt[1];
//                     return 0;}
//
//                 function default1(pt) {
//                   if (pt[0]) return pt[1];
//                     return 1;}
//
//                 function multParts(pt) {
//                   return pt[0] * pt[1];}
//
//                 function addParts(pt) {
//                   return pt[0] + pt[1];}
//
//                 function digitToNum(chr) {
//                   if (chr === '0') return 0;
//                   if (chr === '1') return 1;
//                   if (chr === '2') return 2;
//                   if (chr === '3') return 3;
//                   if (chr === '4') return 4;
//                   if (chr === '5') return 5;
//                   if (chr === '6') return 6;
//                   if (chr === '7') return 7;
//                   if (chr === '8') return 8;
//                   if (chr === '9') return 9;
//                   if (chr === 'A' || chr === 'a') return 10;
//                   if (chr === 'B' || chr === 'b') return 11;
//                   if (chr === 'C' || chr === 'c') return 12;
//                   if (chr === 'D' || chr === 'd') return 13;
//                   if (chr === 'E' || chr === 'e') return 14;
//                   if (chr === 'F' || chr === 'f') return 15;}
//
//                 function digitsToNum(base) {
//                   return function(digits) {
//                            if (digits.length == 0) return 0;
//                            if (digits.length == 1)
//                              return digitToNum(digits[0]);
//                            return digitsToNum(base)(digits.slice(
//                                                       0,
//                                                       digits.length - 1))
//                                   * base
//                                   + digitToNum(digits[digits.length - 1]);};}
//
//                 function fracDigitsToNum(base) {
//                   return function(digits) {
//                     return digitsToNum(1 / base)(digits.reverse()) / base};}
//
//                 function digit(base) {
//                   if (base == 2) return ps.or(ps.string('0'),
//                                               ps.string('1'));
//                   if (base == 8) return ps.or(digit(2),
//                                               ps.string('2'),
//                                               ps.string('3'),
//                                               ps.string('4'),
//                                               ps.string('5'),
//                                               ps.string('6'),
//                                               ps.string('7'));
//                   if (base == 10) return ps.or(digit(8),
//                                                ps.string('8'),
//                                                ps.string('9'));
//                   if (base == 16) return ps.or(digit(10),
//                                                ps.string('A'),
//                                                ps.string('a'),
//                                                ps.string('B'),
//                                                ps.string('b'),
//                                                ps.string('C'),
//                                                ps.string('c'),
//                                                ps.string('D'),
//                                                ps.string('d'),
//                                                ps.string('E'),
//                                                ps.string('e'),
//                                                ps.string('F'),
//                                                ps.string('f'));}
//
//                 function digits(base) {
//                   return ps.many1(digit(base));}
//
//                 var signParser = ps.mapParser(ps.or(ps.string('+'),
//                                                     ps.string('-')),
//                                               function (pt) {
//                                                 return pt === '+' ? 1
//                                                                   : -1;});
//
//                 var numPartParserBase = function(base) {
//                   return ps.or(
//                     ps.mapParser(
//                       digits(base),
//                       digitsToNum(base)),
//                     ps.mapParser(
//                       ps.seq(
//                         ps.mapParser(
//                           ps.opt(
//                             ps.mapParser(
//                               digits(base),
//                               digitsToNum(base))),
//                           default0),
//                         ps.before(
//                           ps.string('.'),
//                           ps.mapParser(
//                             digits(base),
//                             fracDigitsToNum(base)))),
//                       addParts));};
//
//                 var urealParserBase = function(base) {
//                   var prefix = base == 2 ? ps.string('0b') :
//                                base == 8 ? ps.string('0o') :
//                                base == 16 ? ps.string('0x') :
//                                             ps.or(ps.nothing,
//                                                   ps.string('0d'));
//
//                   return ps.before(prefix,
//                                    ps.mapParser(
//                                      ps.seq(
//                                        numPartParserBase(base),
//                                        ps.mapParser(
//                                          ps.opt(
//                                            ps.mapParser(
//                                              ps.before(
//                                                ps.or(ps.string('e'),
//                                                      ps.string('E')),
//                                                ps.mapParser(
//                                                  ps.seq(
//                                                    ps.mapParser(
//                                                      ps.opt(signParser),
//                                                      default1),
//                                                    numPartParserBase(base)),
//                                                  multParts)),
//                                              function (pt) {
//                                                return Math.pow(base, pt);})),
//                                          default1)),
//                                      multParts));}
//
//                 var urealParser = ps.or(urealParserBase(2),
//                                         urealParserBase(8),
//                                         urealParserBase(10),
//                                         urealParserBase(16));
//
//                 var realParser = ps.mapParser(ps.seq(ps.mapParser(
//                                                        ps.opt(signParser),
//                                                        default1),
//                                                      urealParser),
//                                               multParts);
//
//                 var numParser
//                   = ps.or(ps.mapParser(realParser,
//                                        function (pt) {
//                                          return [pt, 0];}),
//                           ps.after(ps.seq(realParser,
//                                           ps.mapParser(
//                                             ps.seq(
//                                               signParser,
//                                               ps.mapParser(
//                                                 ps.opt(urealParser),
//                                                 default1)),
//                                             multParts)),
//                                    ps.string('i')),
//                           ps.mapParser(ps.after(ps.mapParser(
//                                                   ps.seq(
//                                                     ps.mapParser(
//                                                       ps.opt(signParser),
//                                                       default1),
//                                                     ps.mapParser(
//                                                       ps.opt(urealParser),
//                                                       default1)),
//                                                   multParts),
//                                                 ps.string('i')),
//                                        function (pt) {
//                                          return [0, pt];}));

//                 if (maybeStr[1].charAt(0) === '"')
//                   return valObj(strLabel,
//                                 maybeStr[1].slice(1, maybeStr[1].length));

//                 var varParts = maybeStr[1].split(':');
//                 if (varParts.length >= 2) {
//                   return _.reduce(varParts.slice(1, varParts.length),
//                                   function(fn, argument) {
//                                     return apply(fn,
//                                                  valObj(strLabel,
//                                                         argument),
//                                                  env);},
//                                   apply(env,
//                                         valObj(strLabel, varParts[0]),
//                                         env));}

                 if (Var[1][0][1] === '"'.codePointAt(0))
                   return [listLabel, Var[1].slice(1, Var[1].length)];

                 if (stringIs(Var, '+'))
                   return fnOfType
                          ( listLabel
                          , function
                            (args
                            , env
                            ){
                              return _.reduce
                                     ( args
                                     , function (arg0, arg1
                                       ){
                                         if (arg1[0] !== intLabel)
                                           return Null();
                                         return [ intLabel,
                                                  arg0[1]
                                                  + arg1[1]];}
                                     , [intLabel, 0]);});

                 if (stringIs(Var, '-'))
                   return fnOfType
                          ( listLabel
                          , function
                            (args
                            , env
                            ){
                              if (args.length === 0) return [intLabel, -1];

                              if (args[0][0] !== intLabel) return Null();
                              if (args.length === 1)
                                return [intLabel, -args[0][1]];

                              return _.reduce
                                     ( args
                                     , function (arg0, arg1
                                       ){
                                         if (arg1[0] !== intLabel)
                                           return Null();
                                         return [ intLabel,
                                                  arg0[1]
                                                  - arg1[1]];}
                                     , args[0]);});

                 if (stringIs(Var, "environment")) return env;

                 if (stringIs(Var, "print"))
                   return makeFn(function(arg, env) {
                     if (!isString(arg))
                       return Null("print's argument should be a string");

                     process.stdout.write(strVal(arg));

                     return unit;});

                 if (stringIs(Var, "->str"))
                   return makeFn(
                     function toString(arg, env) {
                       if (arg[0] === charLabel)
                         return [ listLabel
                                , [ [charLabel, "'".codePointAt(0)]
                                  , arg
                                  , [charLabel, "'".codePointAt(0)]]];

                       if (arg[0] === intLabel)
                         return [ listLabel
                                , arg[1].toString().split('').map
                                  ( function(chr)
                                    {return [charLabel, chr.codePointAt(0)]})];

                       if (arg[0] === listLabel)
                         return [ listLabel,
                                  [[charLabel, '['.codePointAt(0)]].concat
                                  ( _.reduce
                                    ( arg[1].map
                                      (function (o) {return toString(o, env)[1];})
                                    , function(soFar, elem, idx)
                                      {return idx === 0
                                              ? elem
                                              : soFar.concat
                                                ( [[ charLabel
                                                     , ' '.codePointAt(0)]]
                                                , elem);}
                                    , undefined)
                                  , [[charLabel, ']'.codePointAt(0)]])];

                       return Null();});

                 if (stringIs(Var, "str->unicode"))
                   return makeFn(
                     function(arg, env) {
                       if (!isString(arg)) return Null();
                       return [listLabel,
                               arg[1].map(function(Char) {
                                 return [intLabel, Char[1]];})];});

                 if (stringIs(Var, "unicode->str"))
                   return makeFn(function(arg, env) {
                     if (arg[0] !== listLabel) return Null();

                     return [ listLabel,
                              arg[1].map(function (codepoint) {
                                if (codepoint[0] !== intLabel
                                ) return Null();
                                return [ charLabel
                                       , codepoint[1]];})];});

                 if (stringIs(Var, "length"))
                   return fnOfType
                          ( listLabel
                          , function(arg, env) {return list.length;});

                 if (stringIs(Var, "true"))
                   return True;

                 if (stringIs(Var, "false"))
                   return False;

                 var varStr = strVal(Var);
                 if (/^(\-|\+)?[0-9]+$/.test(varStr))
                   return [intLabel, parseInt(varStr, 10)];

                 return Null
                        ( 'string variable not found in environment: "'
                          + Var[1]);};
               return thisIsDumb();}

             return Null();});
exports.initEnv = initEnv;

Error.stackTraceLimit = Infinity;

