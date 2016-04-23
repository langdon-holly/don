//Dependencies

var ps = require('./parse.js');

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

function valObj(label, data) {
  return [[label, data]];}
exports.valObj = valObj;

function getInterfaceData(o, targetLabel, implementszes, prevInterfaces) {
  if (prevInterfaces === undefined) prevInterfaces = [];

  for (var i = 0; i < o[i].length; i++) {
    var maybeInterfaceData = valInterfaceData(o[i],
                                              targetLabel,
                                              implementszes,
                                              prevInterfaces);
    if (maybeInterfaceData[0]) return maybeInterfaceData[1];} 
  return [false];}
exports.getInterfaceData = getInterfaceData;

function valInterfaceData(val, targetLabel, implementszes, prevInterfaces) {
  if (prevInterfaces.includes(val[0])) return [false];

  if (val[0] === targetLabel) return [true, val[1]];

  prevInterfaces.push(val[0]);

  var applicableImplementszes = implementszes[val[0]];
  for (var i = 0; i < applicableImplementszes.length; i++) {
    var maybeInterfaceData = getInterfaceData(applicableImplementszes
                                                [i]
                                                (val[1]),
                                              targetLabel,
                                              implementszes,
                                              prevInterfaces);
    if (maybeInterfaceData[0]) return maybeInterfaceData[1];}
  return [false];}

function mApply(macro, args, implementszes, env) {
  var transformMaybe = getInterfaceData(macro, macroLabel, implementszes);
  if (!transformMaybe[0]) return [false];

  return transformMaybe[1](args, implementszes, env);}
exports.mApply = mApply;

function apply(fn, args, implementszes, env) {
  var transformMaybe = getInterfaceData(fn, fnLabel, implementszes);
  if (!transformMaybe[0]) return [false];

  return transformMaybe[1](args, implementszes, env);}
exports.apply = apply;

function fnOfTypes(interfaceList, fn) {
  return valObj(fnLabel,
                function(args, implementszes, env) {
                  var argsData = [];
                  for (var i = 0; i < args.length; i++) {
                    var maybeArgData = getInterfaceData(args[i],
                                                        interfaceList[i],
                                                        implementszes);
                    if (!maybeArgData[0]) return [false];
                    argsData.push(maybeArgData[1]);}
                  return fn(argsData, implementszes, env);})}

function parseTreeToAST(pt) {
  var label = pt[0];

  if (label == 'name') return valObj(strLabel, pt[1]);
  if (label == 'form') return valObj(listLabel, pt[1].map(parseTreeToAST));
  if (label == 'list') return valObj(listLabel, [listVar]
                                                .concat(pt[1]
                                                        .map(parseTreeToAST)));
  if (label == 'braceStr') return valObj(ASTBraceStrLabel, pt[1]);

  return Null();}

var macroLabel = {};
exports.macroLabel = macroLabel;

var fnLabel = {};
exports.fnLabel = fnLabel;

var listLabel = {};
exports.listLabel = listLabel;

var numLabel = {};
exports.numLabel = numLabel;

var strLabel = {};
exports.strLabel = strLabel;

var symLabel = {};
exports.symLabel = symLabel;

var ASTBraceStrLabel = {};
exports.ASTBraceStrLabel = ASTBraceStrLabel;

var ASTPrecomputedLabel = {};
exports.ASTPrecomputedLabel = ASTPrecomputedLabel;

var preEvalVar = valObj(symLabel, {});
exports.preEvalVar = preEvalVar;

var listVar = valObj(symLabel, {});
exports.listVar = listVar;

var Null = function() {while (true) {}};
exports.Null = Null;

var eval = valObj(fnLabel, function(args, implementszes, env) {
  var expr = args[0];
  if (args.length >= 2) env = args[1];

  var maybeForm = getInterfaceData(expr,
                                   listLabel,
                                   implementszes);
  if (maybeForm[0]) {
    if (maybeForm[1].length < 1) return Null();
    return mApply(apply(eval,
                        [maybeForm[1][0], env],
                        implementszes,
                        env),
                  maybeForm[1]
                    .slice(1, maybeForm[1].length),
                  implementszes,
                  env);}

  var maybeStr = getInterfaceData(expr,
                                  strLabel,
                                  implementszes);
  var maybeSym = getInterfaceData(expr,
                                  symLabel,
                                  implementszes);
  if (maybeStr[0] || maybeSym[0])
    return apply(env, [expr], implementszes, env);

  var maybeBraceStr = getInterfaceData(expr,
                                         ASTBraceStrLabel,
                                         implementszes);
  if (maybeBraceStr[0])
    return apply(apply(env,
                       [preEvalVar],
                       implementszes,
                       env),
                 [expr],
                 implementszes,
                 env);

  var maybePrecomputed = getInterfaceData(expr,
                                          ASTPrecomputedLabel,
                                          implementszes);
  if (maybePrecomputed[0])
    return maybePrecomputed[1];

  return Null();});
exports.eval = eval;

var initImplementszes = {};
initImplementszes[listLabel]
  = [function(list) {
       return fnOfTypes([numLabel],
                        function(args, implementszes, env) {
                          return list[args[0]];});}];
initImplementszes[fnLabel]
 = [function(fn) {
      return valObj(macroLabel,
                    function(args, implementszes, env) {
                      return fn(args.map(function(arg) {
                                           return apply(eval,
                                                        [arg, env],
                                                        implementszes,
                                                        env);}),
                                implementszes,
                                env);});}];
exports.initImplementszes = initImplementszes;

var initEnv
  = valObj(fnLabel,
           function(args, implementszes, env) {
             var Var = args[0];

             var maybeStr = getInterfaceData(Var,
                                             strLabel,
                                             implementszes);
             if (maybeStr[0]) {
               function default0(pt) {
                 if (pt[0]) return pt[1];
                   return 0;}
               function default1(pt) {
                 if (pt[0]) return pt[1];
                   return 1;}
               function multParts(pt) {
                 return pt[0] * pt[1];}
               function digitToNum(chr) {
                 if (chr === '0') return 0;
                 if (chr === '1') return 1;
                 if (chr === '2') return 2;
                 if (chr === '3') return 3;
                 if (chr === '4') return 4;
                 if (chr === '5') return 5;
                 if (chr === '6') return 6;
                 if (chr === '7') return 7;
                 if (chr === '8') return 8;
                 if (chr === '9') return 9;
                 if (chr === 'A' || chr === 'a') return 10;
                 if (chr === 'B' || chr === 'b') return 11;
                 if (chr === 'C' || chr === 'c') return 12;
                 if (chr === 'D' || chr === 'd') return 13;
                 if (chr === 'E' || chr === 'e') return 14;
                 if (chr === 'F' || chr === 'f') return 15;}
               function digitsToNum(base) {
                 return function(digits) {
                          if (digits.length == 0) return 0;
                          if (digits.length == 1)
                            return digitToNum(digits[0]);
                          return digitsToNum(base)(digits.slice(
                                                     0,
                                                     digits.length - 1))
                                 * base
                                 + digitToNum(digits[digits.length - 1]);};}
               function fracDigitsToNum(base) {
                 return function(digits) {
                   return digitsToNum(1 / base)(reverse(digits)) / base};}
               function digits(base) {
                 if (base == 2) return ps.or(ps.string('0'),
                                             ps.string('1'));
                 if (base == 8) return ps.or(digits(2),
                                             ps.string('2'),
                                             ps.string('3'),
                                             ps.string('4'),
                                             ps.string('5'),
                                             ps.string('6'),
                                             ps.string('7'));
                 if (base == 10) return ps.or(digits(8),
                                              ps.string('8'),
                                              ps.string('9'));
                 if (base == 16) return ps.or(digits(10),
                                              ps.string('A'),
                                              ps.string('a'),
                                              ps.string('B'),
                                              ps.string('b'),
                                              ps.string('C'),
                                              ps.string('c'),
                                              ps.string('D'),
                                              ps.string('d'),
                                              ps.string('E'),
                                              ps.string('e'),
                                              ps.string('F'),
                                              ps.string('f'));}

               var signParser = ps.mapParser(ps.or(ps.string('+'),
                                                   ps.string('-')),
                                             function (pt) {
                                               return pt === '+' ? 1
                                                                : -1;});
               var numPartParserBase = function(base) {
                 return ps.or(
                   ps.mapParser(
                     digits(base),
                     digitsToNum(base)),
                   ps.mapParser(
                     ps.seq(
                       ps.mapParser(
                         ps.opt(
                           ps.mapParser(
                             digits(base),
                             digitsToNum(base))),
                         default0),
                       ps.before(
                         ps.string('.'),
                         ps.mapParser(
                           digits(base),
                           fracDigitsToNum(base)))),
                     function (pt) {
                       return pt[0] + pt[1];}));};
               var numPartParser = ps.or(numPartParserBase(2),
                                         numPartParserBase(8),
                                         numPartParserBase(10),
                                         numPartParserBase(16));
               var urealParser = ps.mapParser(ps.seq(
                                                numPartParser,
                                                ps.mapParser(
                                                  ps.before(
                                                    ps.or(ps.string('e'),
                                                          ps.string('E')),
                                                    ps.mapParser(
                                                      ps.seq(
                                                        ps.mapParser(
                                                          ps.opt(signParser),
                                                          default1),
                                                        numPartParser),
                                                      multParts)),
                                                  function (pt) {
                                                    return 10^pt})),
                                              multParts);
               var realParser = ps.mapParser(ps.seq(ps.mapParser(
                                                      ps.opt(signParser),
                                                      default1),
                                                    urealParser),
                                             multParts);
               var numParser
                 = ps.or(ps.mapParser(realParser,
                                      function (pt) {
                                        return [pt, 0];}),
                         ps.after(ps.seq(realParser,
                                         ps.mapParser(
                                           ps.seq(
                                             signParser,
                                             mapParser(
                                               ps.opt(urealParser),
                                               default1)),
                                           multParts)),
                                  ps.string('i')),
                         ps.mapParser(ps.after(ps.mapParser(
                                                 ps.seq(
                                                   ps.mapParser(
                                                     ps.opt(signParser),
                                                     default1),
                                                   ps.mapParser(
                                                     ps.opt(urealParser),
                                                     default1)),
                                                 multParts),
                                               ps.string('i'))),
                                      function (pt) {
                                        return [0, pt];});

               var maybeNum = numParser(maybeStr[1]);
               if (maybeNum[0]) return valObj(numLabel, maybeNum[1]);

               if (maybeStr[1] === '+')
                 return valObj(fnLabel,
                               function sum(args, implementszes, env) {
                                 if (args.length == 0)
                                   return valObj(numLabel, [0, 0]);
                                 if (args.length == 1) return args[0];
                                 if (args.length == 2) {
                                   var maybeNum0 = getInterfaceData(
                                     args[0],
                                     numLabel,
                                     implementszes);
                                   var maybeNum1 = getInterfaceData(
                                     args[1],
                                     numLabel,
                                     implementszes);
                                   if (!maybeNum0[0] || !maybeNum1[0])
                                     return Null();
                                   return valObj(numLabel,
                                                 [maybeNum0[1][0]
                                                  + maybeNum1[1][0],
                                                  maybeNum0[1][1]
                                                  + maybeNum1[1][1]]);}
                                 return args.reduce(sum);});

               return Null();}

             return Null();});
exports.initEnv = initEnv;

