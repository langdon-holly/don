'use strict'

// Dependencies

; var
    ps = require('list-parsing')
  , parser = require('./don-parse.js')
  , _ = require('lodash')
  , nat = require('./nat.js')

// Stuff

; var exports = module.exports

; function apply(fn, arg)
  { function apply2(fn, arg)
    { var funLabel = fn.type, funData = fn.data
    ; return (
        funLabel === fnLabel
        ? funData(arg)
        : funLabel === listLabel
          ? arg.type !== intLabel
            ? Null("Argument to list must be integer")
            : arg.data < 0 || arg.data >= fn.data.length
              ? Null("Array index out of bounds")
              : funData[arg.data]
          : funLabel === quoteLabel
            ? funData
            : funLabel === callLabel
              ? apply(funData.fnExpr, arg, apply(funData.argExpr, arg))
              : funLabel === symLabel || funLabel === identLabel
                ? apply(arg, fn, arg)
                : Null("Tried to apply a non-function"))}
  ; return _.reduce(arguments, apply2)}
; exports.apply = apply

; function mk(label, data) {return {type: label, data: data}}

; function makeCall(fnExpr, argExpr)
  {return mk(callLabel, {fnExpr: fnExpr, argExpr: argExpr})}

; function fnOfType(type, fn)
  {return (
     makeFn
     ( function(arg)
       { if (arg.type !== type) return Null("typed function received garbage")
       ; return fn(arg.data)}))}

; function makeFn(fn) {return mk(fnLabel, fn)}

//; function constFn(val) {return makeFn(_.constant(val))}

; function quote(val) {return mk(quoteLabel, val)}

; function makeList(vals) {return mk(listLabel, vals)}

; function just(val) {return makeList([True, val])}

; function makeInt(Int) {return mk(intLabel, Int)}

; function makeChar(codepoint) {return mk(charLabel, codepoint)}

; function gensym(debugId) {return mk(symLabel, {sym: debugId})}

; function makeIdent(val) {return mk(identLabel, val)}

; function isTrue(val)
  { var trueVal = gensym('if-true')
  ; return apply(val, makeList([trueVal, gensym('if-false')])) === trueVal}

; function isString(val)
  {return (
     val.type === listLabel
     && _.every(val.data, function(elem) {return elem.type === charLabel}))}

; function strVal(list)
  { if (list.type !== listLabel) return Null()
  ; return (
      list.data.reduce
      ( function(soFar, chr) {return soFar + charToStr(chr)}
      , ''))}

; function stringIs(list, str)
  {return strVal(list) === str}

; function charToStr(Char)
  { if (Char.type !== charLabel)
      return Null("charToStr nonchar: " + strVal(toString(Char)))
  ; return String.fromCodePoint(Char.data)}

; function strToChars(str)
  {return (
     makeList
     ( Array.from(str).map
       (function(chr) {return makeChar(chr.codePointAt(0))})))}

; function eq(val0, val1)
  {return (
     val0.type === val1.type
     &&
       ( val0.data === val1.data
       ||
         val0.type === listLabel
         && val0.data.length == val1.data.length
         && _.every
            ( val0.data
            , function(elem, index) {return eq(elem, val1.data[index])})
       || val0.type === quoteLabel && eq(val0.data, val1.data)
       || val0.type === identLabel && eq(val0.data, val1.data)
       ||
         val0.type === callLabel
         && eq(val0.data.fnExpr, val1.data.fnExpr)
         && eq(val0.data.argExr, val1.data.argExr)))}

; function parseTreeToAST(pt)
  { var label = pt[0]
  ; var data = pt[1]

  ; if (label == 'char') return makeChar(data)
  ; if (label == 'call')
      return (
        _.reduce
        ( data
        , function(applied, arg)
          {return makeCall(applied, parseTreeToAST(arg))}
        , quote(makeFn(function(arg){return arg}))))
  ; if (label == 'bracketed')
      return (
        makeCall
        ( bracketedVar
        , makeFn
          ( function(env)
            {return (
               makeList
               ( data.map
                 ( _.flow
                   ( parseTreeToAST
                   , function(expr) {return apply(expr, env)}))))})))
  ; if (label == 'braced')
      return (
        makeCall
        ( bracedVar
        , makeFn
          ( function(env)
            {return (
               makeList
               ( data.map
                 ( _.flow
                   ( parseTreeToAST
                   , function(expr) {return apply(expr, env)}))))})))
  ; if (label === 'heredoc')
      return quote(makeList(data.map(parseTreeToAST)))

  ; if (label === 'quote') return quote(parseTreeToAST(data))

  ; if (label === 'ident')
      return (
        makeIdent(makeList(data.map(function(chr) {return makeChar(chr)}))))

  ; return Null('unknown parse-tree type "' + label)}

; function parseStr(str)
  { var parsed = parser(str)
  ; return (
      parsed.status === 'match' 
      ? _.assign({ast: parseTreeToAST(parsed.result)}, parsed)
      : parsed)}
; exports.parse = parseStr

; function ttyLog()
  { if (process.stdout.isTTY) console.log.apply(this, arguments)}

; var fnLabel = {label: 'fn'}
; exports.fnLabel = fnLabel

; var listLabel = {label: 'list'}
; exports.listLabel = listLabel

; var intLabel = {label: 'int'}
; exports.intLabel = intLabel

; var charLabel = {label: 'char'}
; exports.charLabel = charLabel

; var symLabel = {label: 'sym'}
; exports.symLabel = symLabel

; var quoteLabel = {label: 'quote'}
; exports.quoteLabel = quoteLabel

; var unitLabel = {label: 'unit'}
; exports.unitLabel = unitLabel
; var unit = mk(unitLabel)
; exports.unit = unit

; var callLabel = {label: 'call'}
; exports.callLabel = callLabel

; var identLabel = {label: 'ident'}
; exports.identLabel = identLabel

; var bracketedVar = gensym('bracketed-var')
; exports.bracketedVar = bracketedVar

; var bracedVar = gensym('braced-var')
; exports.bracedVar = bracedVar

; var Null
  = function()
    { ttyLog.apply(this, arguments)
    ; throw new Error("diverging…")
    ; while (true) {}}
; exports.Null = Null

; var False
  = makeFn
    ( function(consequent)
      {return makeFn(function(alternative) {return alternative})})

; var True
  = makeFn
    ( function(consequent)
      {return makeFn(function(alternative) {return consequent})})

; var nothing = makeList([False])

; var topEval = function(ast, rest)
{ //var calls = []
  //; while (continuing.length > 0)
  ; return apply(ast, initEnv)}
; exports.topEval = topEval

; function toString(arg)
  { var argLabel = arg.type, argData = arg.data
  ; if (argLabel === charLabel)
      return (
        makeList
        ( [ strToChars("'").data[0]
          , arg
          , strToChars("'").data[0]]))

  ; if (argLabel === intLabel)
      return (
        makeList
        ( _.toArray(argData.toString()).map
          ( function(chr) {return strToChars(chr).data[0]})
          .concat(strToChars(' ').data)))

  ; if (argLabel === listLabel)
      return (
        makeList
        ( strToChars('[').data.concat
          ( _.reduce
            ( argData.map
              (function(o) {return toString(o).data})
            , function(soFar, elem, idx)
              {return (
                 idx === 0
                 ? elem
                 : soFar.concat
                   (elem))}
            , [])
          , strToChars(']').data)))

  ; if (argLabel === quoteLabel)
      return makeList([makeChar(34)].concat(toString(argData).data))

  ; if (argLabel === unitLabel)
      return (
        makeList
        ( [117, 110, 105, 116, 32]
          .map(function(chr) {return makeChar(chr)})))

  ; if (argLabel === callLabel)
      return (
        makeList
        ( [makeChar(92)]
          .concat(toString(argData.fnExpr).data)
          .concat(toString(argData.argExpr).data)))

  ; if (argLabel === identLabel)
      return (
        makeList
        ( isString(argData)
          ? strToChars('"|').data.concat
            ( _.flatMap
              ( argData.data
              , function(chr)
                {return (
                   chr.data == 92 || chr.data == 124
                   ? [makeChar(92), chr]
                   : [chr])})
            , strToChars("|").data)
          : strToChars("(ident ").data.concat
            (toString(argData).data, strToChars(")").data)))

  ; if (argLabel === fnLabel) return strToChars("(fn ... )")

  ; if (argLabel === symLabel) return strToChars("(sym ... )")

  ; return Null("->str unknown type:", arg)}

; var initEnv
  = makeFn
    ( function(Var)
      { if (Var.type === symLabel)
        { if (Var === bracketedVar) return quote(makeFn(_.identity))

        ; if (Var === bracedVar)
            return (
              quote
              ( fnOfType
                ( listLabel
                , function(args)
                  { if (args.length % 2 != 0) return Null("Tried to brace oddity")
                  ; var pairs = _.chunk(args, 2)
                  ; return (
                    makeFn
                    ( function(arg)
                      { var toReturn = nothing
                      ; _.forEach
                        ( pairs
                        , function(pair)
                          {return (
                             eq(arg, pair[0])
                             ? (toReturn = just(pair[1]), false)
                             : true)})
                      ; return toReturn}))})))

        ; return Null("symbol variable not found in environment")}

      ; if (Var.type === identLabel && isString(Var.data))
        { var
            vaR = Var
          , thisIsDumb
            = function()
              { var Var = vaR.data

  //            function default0(pt) {
  //              if (pt[0]) return pt[1];
  //                return 0}
  //
  //            function default1(pt) {
  //              if (pt[0]) return pt[1];
  //                return 1}
  //
  //            function multParts(pt) {
  //              return pt[0] * pt[1]}
  //
  //            function addParts(pt) {
  //              return pt[0] + pt[1]}
  //
  //            function digitToNum(chr) {
  //              if (chr === '0') return 0;
  //              if (chr === '1') return 1;
  //              if (chr === '2') return 2;
  //              if (chr === '3') return 3;
  //              if (chr === '4') return 4;
  //              if (chr === '5') return 5;
  //              if (chr === '6') return 6;
  //              if (chr === '7') return 7;
  //              if (chr === '8') return 8;
  //              if (chr === '9') return 9;
  //              if (chr === 'A' || chr === 'a') return 10;
  //              if (chr === 'B' || chr === 'b') return 11;
  //              if (chr === 'C' || chr === 'c') return 12;
  //              if (chr === 'D' || chr === 'd') return 13;
  //              if (chr === 'E' || chr === 'e') return 14;
  //              if (chr === 'F' || chr === 'f') return 15}
  //
  //            function digitsToNum(base) {
  //              return function(digits) {
  //                       if (digits.length == 0) return 0;
  //                       if (digits.length == 1)
  //                         return digitToNum(digits[0]);
  //                       return digitsToNum(base)(digits.slice(
  //                                                  0,
  //                                                  digits.length - 1))
  //                              * base
  //                              + digitToNum(digits[digits.length - 1])}}
  //
  //            function fracDigitsToNum(base) {
  //              return function(digits) {
  //                return digitsToNum(1 / base)(digits.reverse()) / base}}
  //
  //            function digit(base) {
  //              if (base == 2) return ps.or(ps.string('0'),
  //                                          ps.string('1'));
  //              if (base == 8) return ps.or(digit(2),
  //                                          ps.string('2'),
  //                                          ps.string('3'),
  //                                          ps.string('4'),
  //                                          ps.string('5'),
  //                                          ps.string('6'),
  //                                          ps.string('7'));
  //              if (base == 10) return ps.or(digit(8),
  //                                           ps.string('8'),
  //                                           ps.string('9'));
  //              if (base == 16) return ps.or(digit(10),
  //                                           ps.string('A'),
  //                                           ps.string('a'),
  //                                           ps.string('B'),
  //                                           ps.string('b'),
  //                                           ps.string('C'),
  //                                           ps.string('c'),
  //                                           ps.string('D'),
  //                                           ps.string('d'),
  //                                           ps.string('E'),
  //                                           ps.string('e'),
  //                                           ps.string('F'),
  //                                           ps.string('f'))}
  //
  //            function digits(base) {
  //              return ps.many1(digit(base))}
  //
  //            var signParser = ps.mapParser(ps.or(ps.string('+'),
  //                                                ps.string('-')),
  //                                          function (pt) {
  //                                            return pt === '+' ? 1
  //                                                              : -1});
  //
  //            var numPartParserBase = function(base) {
  //              return ps.or(
  //                ps.mapParser(
  //                  digits(base),
  //                  digitsToNum(base)),
  //                ps.mapParser(
  //                  ps.seq(
  //                    ps.mapParser(
  //                      ps.opt(
  //                        ps.mapParser(
  //                          digits(base),
  //                          digitsToNum(base))),
  //                      default0),
  //                    ps.before(
  //                      ps.string('.'),
  //                      ps.mapParser(
  //                        digits(base),
  //                        fracDigitsToNum(base)))),
  //                  addParts))};
  //
  //            var urealParserBase = function(base) {
  //              var prefix = base == 2 ? ps.string('0b') :
  //                           base == 8 ? ps.string('0o') :
  //                           base == 16 ? ps.string('0x') :
  //                                        ps.or(ps.nothing,
  //                                              ps.string('0d'));
  //
  //              return ps.before(prefix,
  //                               ps.mapParser(
  //                                 ps.seq(
  //                                   numPartParserBase(base),
  //                                   ps.mapParser(
  //                                     ps.opt(
  //                                       ps.mapParser(
  //                                         ps.before(
  //                                           ps.or(ps.string('e'),
  //                                                 ps.string('E')),
  //                                           ps.mapParser(
  //                                             ps.seq(
  //                                               ps.mapParser(
  //                                                 ps.opt(signParser),
  //                                                 default1),
  //                                               numPartParserBase(base)),
  //                                             multParts)),
  //                                         function (pt) {
  //                                           return Math.pow(base, pt)})),
  //                                     default1)),
  //                                 multParts))}
  //
  //            var urealParser = ps.or(urealParserBase(2),
  //                                    urealParserBase(8),
  //                                    urealParserBase(10),
  //                                    urealParserBase(16));
  //
  //            var realParser = ps.mapParser(ps.seq(ps.mapParser(
  //                                                   ps.opt(signParser),
  //                                                   default1),
  //                                                 urealParser),
  //                                          multParts);
  //
  //            var numParser
  //              = ps.or(ps.mapParser(realParser,
  //                                   function (pt) {
  //                                     return [pt, 0]}),
  //                      ps.after(ps.seq(realParser,
  //                                      ps.mapParser(
  //                                        ps.seq(
  //                                          signParser,
  //                                          ps.mapParser(
  //                                            ps.opt(urealParser),
  //                                            default1)),
  //                                        multParts)),
  //                               ps.string('i')),
  //                      ps.mapParser(ps.after(ps.mapParser(
  //                                              ps.seq(
  //                                                ps.mapParser(
  //                                                  ps.opt(signParser),
  //                                                  default1),
  //                                                ps.mapParser(
  //                                                  ps.opt(urealParser),
  //                                                  default1)),
  //                                              multParts),
  //                                            ps.string('i')),
  //                                   function (pt) {
  //                                     return [0, pt]}));

  //            if (maybeStr[1].charAt(0) === '"')
  //              return valObj(strLabel,
  //                            maybeStr[1].slice(1, maybeStr[1].length));

  //            var varParts = maybeStr[1].split(':');
  //            if (varParts.length >= 2) {
  //              return _.reduce(varParts.slice(1, varParts.length),
  //                              function(fn, argument) {
  //                                return apply(fn,
  //                                             valObj(strLabel,
  //                                                    argument))},
  //                              apply(env,
  //                                    valObj(strLabel, varParts[0])))}

              ; if (stringIs(Var, 'fn'))
                  return (
                    makeFn
                    ( function(env)
                      {return (
                         makeFn
                         ( function(param)
                           {return (
                              makeFn
                              ( function(body)
                                {return (
                                   makeFn
                                   ( function(arg)
                                     {return (
                                        apply
                                        ( body
                                        , makeFn
                                          ( function(Var)
                                            {return (
                                               eq(Var, param)
                                               ? quote(arg)
                                               : apply(env, Var))})))}))}))}))}))

              ; if (stringIs(Var, '+'))
                  return (
                    quote
                    ( fnOfType
                      ( listLabel
                      , function (args)
                        {return (
                           _.reduce
                           ( args
                           , function (arg0, arg1)
                             { if (arg1.type !== intLabel) return Null()
                             ; return makeInt(arg0.data + arg1.data)}
                           , makeInt(0)))})))

              ; if (stringIs(Var, '-'))
                  return (
                    quote
                    ( fnOfType
                      ( listLabel
                      , function (args)
                        { if (args.length === 0) return makeInt(-1)

                        ; if (args[0].type !== intLabel) return Null()
                        ; if (args.length === 1) return makeInt(-args[0].data)

                        ; return (
                            _.reduce
                            ( args
                            , function (arg0, arg1)
                              { if (arg1.type !== intLabel) return Null()
                              ; return makeInt(arg0.data - arg1.data)}))})))

              ; if (stringIs(Var, '<'))
                  return (
                    quote
                    ( fnOfType
                      ( intLabel
                      , function(arg0)
                        {return (
                           fnOfType
                           ( intLabel
                           , function(arg1)
                             {return arg0 < arg1 ? True : False}))})))

              ; if (stringIs(Var, '='))
                  return (
                    quote
                    ( makeFn
                      ( function(arg0)
                        {return (
                           makeFn
                           ( function(arg1)
                             {return eq(arg0, arg1) ? True : False}))})))

              ; if (stringIs(Var, "env")) return makeFn(function(env) {return env})

              ; if (stringIs(Var, "print"))
                  return (
                    quote
                    ( makeFn
                      ( function(arg)
                        {return (
                           isString(arg)
                           ? process.stdout.write(strVal(arg))
                           : Null('Tried to print nonstring')
                           , unit)})))

              ; if (stringIs(Var, "say"))
                  return (
                    quote
                    ( makeFn
                      ( function(arg)
                        {return (
                           process.stdout.write(strVal(toString(arg)))
                           , unit)})))

              ; if (stringIs(Var, "->str")) return quote(makeFn(toString))

              ; if (stringIs(Var, "str->unicode"))
                  return (
                    quote
                    ( makeFn
                      ( function(arg)
                        { if (!isString(arg)) return Null()
                        ; return (
                            makeList
                            ( arg.data.map
                              ( function(Char)
                                {return makeInt(Char.data)})))})))

              ; if (stringIs(Var, "unicode->str"))
                  return (
                    quote
                    ( makeFn
                      ( function(arg)
                        { if (arg.type !== listLabel) return Null()

                        ; return (
                            makeList
                            ( arg.data.map
                              ( function (codepoint)
                                { if (codepoint.type !== intLabel) return Null()
                                ; return makeChar(codepoint.data)})))})))

              ; if (stringIs(Var, "length"))
                  return (
                    quote
                    ( fnOfType
                      ( listLabel
                      , function(arg) {return makeInt(arg.length)})))

              ; if (stringIs(Var, "->list"))
                  return (
                    quote
                    ( fnOfType
                      ( fnLabel
                      , function(fn)
                        {return (
                           fnOfType
                           ( intLabel
                           , function(length)
                             { if (length < 0) return Null()

                             ; var toReturn = []
                             ; for (var i = 0; i < length; i++)
                                 {toReturn.push(fn(makeInt(i)))}
                             ; return makeList(toReturn)}))})))

              ; if (stringIs(Var, "true")) return quote(True)

              ; if (stringIs(Var, "false")) return quote(False)

              ; if (stringIs(Var, "unit")) return quote(unit)

              ; if (Var.data[0].data == '"'.codePointAt(0))
                  return quote(makeList(Var.data.slice(1)));

              ; var varStr = strVal(Var)
              ; if (/^(\-|\+)?[0-9]+$/.test(varStr))
                  return quote(makeInt(parseInt(varStr, 10)))

              ; return (
                  Null
                  ( 'string variable not found in environment: "'
                    + strVal(Var)))}
        ; return thisIsDumb()}

        return Null("unknown variable: " + strVal(toString(Var)))})
; exports.initEnv = initEnv

; Error.stackTraceLimit = Infinity

