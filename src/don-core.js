'use strict';

// Dependencies

var ps = require('./parse.js');
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

//function valObj(label, data) {
//  return [[label, data]];}
//exports.valObj = valObj;

//function getInterfaceData(o, targetLabel, implementszes, prevInterfaces) {
//  if (prevInterfaces === undefined) prevInterfaces = [];
//
//  for (var i = 0; i < o.length; i++) {
//    var maybeInterfaceData = valInterfaceData(o[i],
//                                              targetLabel,
//                                              implementszes,
//                                              prevInterfaces);
//    if (maybeInterfaceData[0]) return maybeInterfaceData;} 
//  return [false];}
//exports.getInterfaceData = getInterfaceData;
//
//function valInterfaceData(val, targetLabel, implementszes, prevInterfaces) {
//  if (prevInterfaces.includes(val[0])) return [false];
//
//  if (val[0] === targetLabel) return [true, val[1]];
//
//  prevInterfaces.push(val[0]);
//
//  var applicableImplementszes = implementszes(val[0]);
//  for (var i = 0; i < applicableImplementszes.length; i++) {
//    var maybeInterfaceData = getInterfaceData(applicableImplementszes
//                                                [i]
//                                                (val[1]),
//                                              targetLabel,
//                                              implementszes,
//                                              prevInterfaces);
//    if (maybeInterfaceData[0]) return maybeInterfaceData;}
//  return [false];}

function mApply(macro, args, env) {
  if (macro[0] === macroLabel)
    return macro[1](args, env);

  args = args.map(function (arg) {
    return apply(Eval, [arg, env], env);});

  if (macro[0] === fnLabel)
    return macro[1](args, env);

  if (macro[0] === listLabel) {
    if (args.length != 1)
      return Null("Lists are unary");
    if (args[0][0] !== intLabel)
      return Null("Argument to list must be integer");

    return macro[1][args[0][1]];}

  return Null("Tried to macro-apply a non-macro");}
exports.mApply = mApply;

function apply(fn, args, env) {
  return fn[0] === macroLabel ? mApply(fn,
                                       args.map(function (arg) {
                                         return [ASTPrecomputedLabel, arg];}),
                                       env)
         : fn[0] === fnLabel ? fn[1](args, env)
         : fn[0] === listLabel ? function(){
           if (args.length != 1)
             return Null("Lists are unary");
           if (args[0][0] !== intLabel)
             return Null("Argument to list must be integer");

           return fn[1][args[0][1]];}()
         : Null("Tried to apply a non-macro");}
exports.apply = apply;

function fnOfTypes(interfaceList, fn) {
  return makeFn(function(args, env) {
                  var argsData = [];
                  for (var i = 0; i < args.length; i++) {
                    if (args[i][0] !== interfaceList[i]) return Null();
                    argsData.push(args[i][1]);}
                  return fn(argsData, env);})}

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
  return apply(val, [trueVal, [symLabel, {}]], env) === trueVal;}

function parseTreeToAST(pt) {
  var label = pt[0];

  if (label == 'name') return [strLabel, pt[1]];
  if (label == 'form') return [listLabel, pt[1].map(parseTreeToAST)];
  if (label == 'list') return [listLabel,
                               [listVar].concat(pt[1]
                                                  .map(parseTreeToAST))];
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
                              elem[0] === 'form'
                              ? ['form', parseTreeToAST(elem)]
                              : elem];})]]]];
  if (label === 'heredoc') return [ASTPrecomputedLabel, [strLabel, pt[1]]];

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

var strLabel = {};
exports.strLabel = strLabel;

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

var ASTBraceStrVar = [symLabel, {}];
exports.ASTBraceStrVar = ASTBraceStrVar;

var Null = function() {
  ttyLog.apply(this, arguments);
  throw new Error("diverging…");
  while (true) {}};
exports.Null = Null;

var nothing = makeList(false);

var True = [macroLabel, function(args, env) {
  if (args.length != 2) return Null("true is binary");

  return apply(Eval,
               [args[0], env],
               env);}];

var Eval = makeFn(function(args, env) {
  if (args.length < 1 || args.length > 2) return Null();

  var expr = args[0];
  if (args.length >= 2) env = args[1];

  if (expr[0] === listLabel) {
    if (expr[1].length < 1) return Null();
    return mApply(apply(Eval,
                        [expr[1][0], env],
                        env),
                  expr[1]
                    .slice(1, expr[1].length),
                  env);}

  if (expr[0] === strLabel || expr[0] === symLabel) {
    return apply(env, [expr], env);
    return Null("variable not found in environment");}

  if (expr[0] === ASTPrecomputedLabel)
    return expr[1];

  return Null();});
exports.Eval = Eval;

var topEval = function(ast) {
  return apply(Eval, [ast], initEnv);}
exports.topEval = topEval;

//var initImplementszes = function(Interface) {
//if (Interface === listLabel)
//  return [function(list) {
//    return fnOfTypes([intLabel],
//                     function(args, env) {
//                       return list[args];});}];
//if (Interface === fnLabel)
// return [function(fn) {
//   return valObj(macroLabel,
//                 function(args, env) {
//                   return fn(args.map(function(arg) {
//                                        return apply(Eval,
//                                                     [arg, env],
//                                                     env);}),
//                             env);});}];
//return [];}
//exports.initImplementszes = initImplementszes;

var initEnv
  = makeFn(function(args, env) {
             var Var = args[0];

             if (Var[0] === strLabel) {
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
//                                                  [valObj(strLabel,
//                                                          argument)],
//                                                  env);},
//                                   apply(env,
//                                         [valObj(strLabel, varParts[0])],
//                                         env));}

                 if (Var[1][0] === '"')
                   return [strLabel, Var[1].slice(1, Var[1].length)];

                 if (Var[1] === '+')
                   return makeFn(function sum(args, env) {
                     args.map(function (arg) {
                       if (arg[0] !== intLabel) return Null();});
                     if (args.length == 0)
                       return [intLabel, 0];
                     if (args.length == 1) return args[0];
                     if (args.length == 2) {
                       return [intLabel,
                               args[0][1]
                               + args[1][1]];}
                     return args.reduce(sum);});

                 if (Var[1] === '-')
                   return makeFn(function difference(args, env) {
                     args.map(function (arg) {
                       if (arg[0] !== intLabel) return Null();});
                     if (args.length == 0)
                       return [intLabel, 0];
                     if (args.length == 1)
                       return [intLabel, -args[0][1]];
                     if (args.length == 2) {
                       return [intLabel,
                                     args[0][1]
                                     - args[1][1]];}
                     return args.reduce(sum);});

                 if (Var[1] === "environment") return env;

                 if (Var[1] === "print")
                   return makeFn(function(args, env) {
                     if (args.length != 1)
                       return Null("print is unary");

                     if (args[0][0] !== strLabel)
                       return Null(
                         "print's argument should be a string");

                     process.stdout.write(args[0][1]);

                     return unit;});

                 if (Var[1] === "->str")
                   return makeFn(
                     function toString(args, env) {
                       if (args.length != 1) return Null();

                       var arg = args[0];

                       if (arg[0] === strLabel)
                         return [strLabel, '{' + arg[1] + '}'];

                       if (arg[0] === intLabel)
                         return [strLabel, arg[1].toString()];

                       if (arg[0] === listLabel)
                         return [strLabel,
                                 '['
                                 + arg[1].map(function (o) {
                                   return toString([o],
                                                   env)[1];}).join(' ')
                                 + ']'];

                       return Null();});

                 if (Var[1] === "str->unicode")
                   return makeFn(
                     function(args, env) {
                       if (args.length != 1) return Null();
                       if (args[0][0] !== strLabel) return Null();

                       var codepoints = [];
                       for (var i = 0; i < args[0][1].length; i++) {
                         var codepoint = args[0][1].codePointAt(i);
                         if (codepoint !== undefined)
                           codepoints.push(codepoint);}

                       return [listLabel,
                               codepoints.map(function(codepoint) {
                                 return [intLabel, codepoint];})];});

                 if (Var[1] === "unicode->str")
                   return makeFn(function(args, env) {
                     if (args.length != 1) return Null();
                     if (args[0][0] !== listLabel) return Null();

                     return [strLabel,
                             String
                             .fromCodePoint
                             .apply(this,
                                    args[0][1].map(function (codepoint) {
                                      return codepoint[1];}))];});

                 if (Var[1] === "length")
                   return makeFn(function(args, env) {
                     if (args.length != 1) return Null();

                     if (args[0][0] === listLabel)
                       return args[0][1].length;
                     return Null("invalid argument for length");});

                 if (Var[1] === "true")
                   return True;

                 if (Var[1] === "false")
                   return [macroLabel, function(args, env) {
                     if (args.length != 2) return Null("false is binary");

                     return apply(Eval,
                                  [args[1], env],
                                  env);}];

                 if (/^(\-|\+)?[0-9]+$/.test(Var[1]))
                   return [intLabel, parseInt(Var[1], 10)];

                 return Null
                        ( 'string variable not found in environment: "'
                          + Var[1]);};
               return thisIsDumb();}

             if (Var[0] === symLabel) {

               if (Var === preEvalVar)
                 return makeFn(function(args, env) {
                   if (args.length != 1) return Null();

                   if (args[0][0] !== listLabel) return Null();

                   return [ strLabel,
                            _
                            .reduce
                            ( _
                              .map
                              ( args[0][1],
                                function(elem) {
                                  if (elem[0] !== ASTBraceStrElemLabel)
                                    return Null();
                                  return elem[1];}),
                              function(soFar, elem) {
                                var toAppend;
                                if (elem[0] === 'str') toAppend = elem[1];
                                else if (elem[0] === 'escape') {
                                  var
                                    evaled
                                  = apply
                                    (Eval, [[strLabel, elem[1]], env]);
                                  if (evaled[0] !== strLabel) return Null();
                                  toAppend = evaled[1];}
                                else if (elem[0] === 'form') {
                                  var
                                    evaled
                                  = apply
                                    (Eval, [elem[1], env]);
                                  if (evaled[0] !== strLabel) return Null();
                                  toAppend = evaled[1];}
                                else return Null();
                                return soFar + toAppend;},
                              '')];});

               if (Var === listVar)
                 return makeFn(function(args, env) {
                   return [listLabel, args];});

               if (Var === ASTBraceStrVar)
                 return makeFn(function(args, env) {
                   if (args[0][0] !== listLabel) return Null();
                   return [strLabel, args[0][1].join('')];});

               if (Var === braceStrEvalVar) return Eval;

               return Null();}

             return Null();});
exports.initEnv = initEnv;

Error.stackTraceLimit = Infinity;

