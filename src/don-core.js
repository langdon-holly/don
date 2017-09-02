'use strict'

// Dependencies

; const
    fs = require('fs')
  , util = require('util')

  , _ = require('lodash')
  //, bigInt = require('big-integer')

  , ps = require('list-parsing')
  , parser = require('./don-parse.js')

// Utility
; const
    debug = true
  , log = (...args) => (debug ? console.log(...args) : undefined, _.last(args))

// Stuff

; exports = module.exports

; function apply(fn, ...args)
  { return (
      _.reduce
      ( args
      , (fn, arg) =>
          { const
              {type: funLabel, data: funData} = fn
            , {type: argLabel, data: argData} = arg
          ; return (
              funLabel === fnLabel
              ? funData(arg)
              : funLabel === listLabel
                ? argLabel !== intLabel
                  ? Null("Argument to list must be integer")
                  : argData < 0 || argData >= funData.length
                    ? Null("Array index out of bounds")
                    : funData[argData]
                : funLabel === quoteLabel
                  ? funData
                  : funLabel === callLabel
                    ? apply
                      (apply(funData.fnExpr, arg), apply(funData.argExpr, arg))
                    : funLabel === identLabel
                      ? apply(arg, fn, arg)
                      : funLabel === boolLabel
                        ? funData
                          ? makeFn(_.constant(arg))
                          : makeFn(_.identity)
                        : Null("Tried to apply a non-function"))}
      , fn))}
; exports.apply = apply

; function mk(label, data) {return {type: label, data: data}}

; function makeCall(fnExpr, argExpr) {return mk(callLabel, {fnExpr, argExpr})}

; function fnOfType(type, fn)
  { return (
      makeFn
      ( arg =>
        { if (arg.type !== type) return Null("typed function received garbage")
        ; return fn(arg.data)}))}

; function makeFn(fn) {return mk(fnLabel, fn)}

//; function constFn(val) {return makeFn(_.constant(val))}

; function quote(val) {return mk(quoteLabel, val)}

; function makeList(vals) {return mk(listLabel, vals)}

; function just(val) {return mk(maybeLabel, {is: true, val})}

; function makeInt(Int) {return mk(intLabel, Int)}

; function makeChar(codepoint) {return mk(charLabel, codepoint)}

; function gensym(debugId) {return mk(symLabel, {sym: debugId})}

; function makeIdent(val) {return mk(identLabel, val)}

; function makeBool(val) {return mk(boolLabel, val)}

//; function isTrue(val)
//  { const trueVal = makeIdent(gensym('if-true'))
//  ; return (
//      apply(val, makeList([trueVal, makeIdent(gensym('if-false'))]))
//      === trueVal)}

; function isString(val)
  { return (
      val.type === listLabel
      && _.every(val.data, elem => elem.type === charLabel))}

; function strVal(list)
  { if (list.type !== listLabel) return Null("Tried to strVal nonlist")
  ; return list.data.reduce((soFar, chr) => soFar + charToStr(chr), '')}

; function stringIs(list, str)
  {return strVal(list) === str}

; function charToStr(Char)
  { if (Char.type !== charLabel)
      return Null("charToStr nonchar: " + strVal(toString(Char)))
  ; return String.fromCodePoint(Char.data)}

; function strToChar(chr) {return makeChar(chr.codePointAt(0))}

; function strToChars(str)
  {return makeList(Array.from(str).map(strToChar))}

; function eq(val0, val1)
  { return (
      val0.type === val1.type
      &&
        ( val0.data === val1.data
        ||
          val0.type === listLabel
          && val0.data.length == val1.data.length
          && _.every(val0.data, (elem, index) => eq(elem, val1.data[index]))
        || val0.type === quoteLabel && eq(val0.data, val1.data)
        || val0.type === identLabel && eq(val0.data, val1.data)
        ||
          val0.type === callLabel
          && eq(val0.data.fnExpr, val1.data.fnExpr)
          && eq(val0.data.argExr, val1.data.argExr)))}

; function parseTreeToAST(pt)
  { const label = pt[0]
  ; const data = pt[1]

  ; if (label == 'char') return quote(makeChar(data))
  ; if (label == 'call')
      return (
        data.length === 0
        ? quote(makeFn(_.identity))
        : _.reduce(_.map(data, parseTreeToAST), makeCall))
  ; if (label == 'bracketed')
      return (
        makeCall
        ( bracketedVar
        , makeFn
          ( env =>
              makeList
              (data.map(_.flow(parseTreeToAST, expr => apply(expr, env)))))))
  ; if (label == 'braced')
      return (
        makeCall
        ( bracedVar
        , makeFn
          ( env =>
              makeList
              (data.map(_.flow(parseTreeToAST, expr => apply(expr, env)))))))
  ; if (label === 'heredoc')
      return quote(makeList(data.map(parseTreeToAST)))

  ; if (label === 'quote') return quote(parseTreeToAST(data))

  ; if (label === 'ident') return makeIdent(makeList(data.map(makeChar)))

  ; return Null('unknown parse-tree type "' + label)}

; function parseStr(str)
  { const parsed = parser(str)
  ; return (
      parsed.status === 'match' 
      ? _.assign({ast: parseTreeToAST(parsed.result)}, parsed)
      : parsed)}

//; function ttyLog()
//  { if (process.stdout.isTTY) console.log(...arguments)}

; const fnLabel = {label: 'fn'}
; exports.fnLabel = fnLabel

; const listLabel = {label: 'list'}
; exports.listLabel = listLabel

; const intLabel = {label: 'int'}
; exports.intLabel = intLabel

; const charLabel = {label: 'char'}
; exports.charLabel = charLabel

; const symLabel = {label: 'sym'}
; exports.symLabel = symLabel

; const quoteLabel = {label: 'quote'}
; exports.quoteLabel = quoteLabel

; const unitLabel = {label: 'unit'}
; exports.unitLabel = unitLabel
; const unit = mk(unitLabel)
; exports.unit = unit

; const callLabel = {label: 'call'}
; exports.callLabel = callLabel

; const identLabel = {label: 'ident'}
; exports.identLabel = identLabel

; const maybeLabel = {label: 'maybe'}
; exports.maybeLabel = maybeLabel

; const boolLabel = {label: 'bool'}
; exports.boolLabel = boolLabel

; const bracketedVarSym = gensym('bracketed-var')
; const bracketedVar = makeIdent(bracketedVarSym)
; exports.bracketedVar = bracketedVar

; const bracedVarSym = gensym('braced-var')
; const bracedVar = makeIdent(bracedVarSym)
; exports.bracedVar = bracedVar

; const Null
  = (...args) => {throw new Error("Diverging: " + util.format(...args))}
; exports.Null = Null

; const False = makeBool(false)

; const True = makeBool(true)

; const nothing = mk(maybeLabel, {is: false})

; const readFile = filename => fs.readFileSync(filename, 'utf8')

; const indexToLineColumn
  = (index, string) =>
      { const arr = Array.from(string)
      ; let line = 0, col = 0
      ; for (let i = 0; i < arr.length; i++)
        { if (i == index)
            return {line0: line, col0: col, line1: ++line, col1: ++col}
        ; if (arr[i] === '\n') line++, col = 0
        ; else col++}
        throw new RangeError
                  ("indexToLineColumn: index=" + index + " is out of bounds")}

; const parseFile
  = data =>
      { const parsed = parseStr(data)

      ; if (parsed.status === 'match')
          return { success: true, ast: parsed.ast, rest: parsed.rest}
      ; else if (parsed.status === 'eof')
          return (
            { success: false
            , error
              : filename =>
                  "Syntax error: "
                  + filename
                  + " should have at least "
                  + (Array.from(data).length + 1)
                  + " codepoints"})
      ; else
        { const errAt = parsed.index
        ; if (errAt == 0)
            return {success: false, error: _.constant("Error in the syntax")}
        ; else
          { const lineCol = indexToLineColumn(errAt - 1, data)
          ; return (
              { success: false
              , error
                : filename =>
                    "Syntax error at "
                    + filename
                    + " "
                    + lineCol.line1
                    + ","
                    + lineCol.col1
                    + ":\n"
                    + data.split('\n')[lineCol.line0]
                    + "\n"
                    + " ".repeat(lineCol.col0)
                    + "^"})}

        //; const trace = parsed.trace
        //; _.forEachRight(trace, function(frame) {console.log("in", frame[0])})
        //; console.log(parsed.parser)
        }}
; exports.parse = parseFile

; const topEval
  = (ast, rest) =>
      ( quotedSourceData =>
          apply
          ( ast
          , makeFn
            ( Var =>
                eq(Var, makeIdent(strToChars('source-data')))
                ? quotedSourceData
                : apply(initEnv, Var))))
      (quote(strToChars(rest)))
//let calls = []
//; while (continuing.length > 0)
; exports.topEval = topEval

; function escInIdent(charArr)
  { return (
      _.flatMap
      ( charArr
      , chr =>
          chr.data == 92 || chr.data == 124
          ? [makeChar(92), chr]
          : [chr]))}

; function toString(arg)
  { const argLabel = arg.type, argData = arg.data
  ; if (argLabel === charLabel) return makeList([strToChar("`"), arg])

  ; if (argLabel === intLabel) return strToChars(argData.toString() + ' ')

  ; if (argLabel === listLabel)
      return (
        makeList
        ( argData.length > 0 && isString(arg)
          ? strToChars('|"').data.concat(escInIdent(argData), [strToChar('|')])
          : [strToChar('[')].concat
            ( _.reduce
              ( argData.map(o => toString(o).data)
              , (soFar, elem) => soFar.concat(elem)
              , [])
            , [strToChar(']')])))

  ; if (argLabel === quoteLabel)
      return makeList([makeChar(34)].concat(toString(argData).data))

  ; if (argLabel === unitLabel) return strToChars('unit ')

  ; if (argLabel === callLabel)
      return (
        makeList
        ( strToChars("(make-call ").data.concat
          ( toString(argData.fnExpr).data
          , toString(argData.argExpr).data
          , [strToChar(")")])))

  ; if (argLabel === identLabel)
      return (
        makeList
        ( isString(argData)
          ? strToChars('"|').data.concat
            (escInIdent(argData.data), [strToChar("|")]) 
          : strToChars("(make-ident ").data.concat
            (toString(argData).data, [strToChar(")")])))

  ; if (argLabel === fnLabel) return strToChars("(fn ... )")

  ; if (argLabel === symLabel)
      return (
        makeList
        ( strToChars("(sym ").data.concat
          (toString(strToChars(argData.sym)).data , [strToChar(")")])))

  ; if (argLabel === maybeLabel)
      return (
        argData.is
        ? makeList
          ( strToChars("(just ").data.concat
            (toString(argData.val).data, [strToChar(")")]))
        : strToChars("nothing "))

  ; if (argLabel === boolLabel)
      return argData ? strToChars("true ") : strToChars("false ")

  ; return Null("->str unknown type:", arg)}

; const initEnv
  = fnOfType
    ( identLabel
    , varKey =>
        varKey.type === symLabel
        ? varKey === bracketedVarSym ? quote(makeFn(_.identity))

          : varKey === bracedVarSym
            ? quote
              ( fnOfType
                ( listLabel
                , args =>
                  { if (args.length % 2 != 0) return Null("Tried to brace oddity")
                  ; const pairs = _.chunk(args, 2)
                  ; return (
                    makeFn
                    ( arg =>
                      { let toReturn = nothing
                      ; _.forEach
                        ( pairs
                        , pair =>
                            eq(arg, pair[0])
                            ? (toReturn = just(pair[1]), false)
                            : true)
                      ; return toReturn}))}))

          : Null("symbol variable not found in environment")

        : isString(varKey)

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

          ? stringIs(varKey, 'fn')
            ? makeFn
              ( env =>
                  makeFn
                  ( param =>
                      makeFn
                      ( body =>
                          makeFn
                          ( arg =>
                              apply
                              ( body
                              , makeFn
                                ( varKey =>
                                    eq(varKey, param)
                                    ? quote(arg)
                                    : apply(env, varKey)))))))

            : stringIs(varKey, '+')
              ? quote
                ( fnOfType
                  ( listLabel
                  , args =>
                      _.reduce
                      ( args
                      , (arg0, arg1) =>
                        { if (arg1.type !== intLabel) return Null()
                        ; return makeInt(arg0.data + arg1.data)}
                      , makeInt(0))))

            : stringIs(varKey, '-')
              ? quote
                ( fnOfType
                  ( listLabel
                  , args =>
                    { if (args.length === 0) return makeInt(-1)

                    ; if (args[0].type !== intLabel) return Null()
                    ; if (args.length === 1) return makeInt(-args[0].data)

                    ; return (
                        _.reduce
                        ( args
                        , (arg0, arg1) =>
                          { if (arg1.type !== intLabel) return Null()
                          ; return makeInt(arg0.data - arg1.data)}))}))

            : stringIs(varKey, '<')
              ? quote
                ( fnOfType
                  ( intLabel
                  , arg0 =>
                      fnOfType
                      ( intLabel
                      , makeBool(arg1 => arg0 < arg1))))

            : stringIs(varKey, '=')
              ? quote
                (makeFn(arg0 => makeFn(arg1 => makeBool(eq(arg0, arg1)))))

            : stringIs(varKey, "env") ? makeFn(_.identity)

            : stringIs(varKey, "init-env") ? quote(initEnv)

            : stringIs(varKey, "print")
              ? quote
                ( makeFn
                  ( arg =>
                      ( isString(arg)
                        ? process.stdout.write(strVal(arg))
                        : Null('Tried to print nonstring')
                        , unit)))

            : stringIs(varKey, "say")
              ? quote
                ( makeFn
                  ( _.flow
                    ( toString
                    , strVal
                    , process.stdout.write.bind(process.stdout)
                    , _.constant(unit))))

            : stringIs(varKey, "->str") ? quote(makeFn(toString))

            : stringIs(varKey, "char->unicode")
              ? quote(fnOfType(charLabel, makeInt))

            : stringIs(varKey, "unicode->char")
              ? quote(fnOfType(intLabel, makeChar))

            : stringIs(varKey, "length")
              ? quote(fnOfType(listLabel, arg => makeInt(arg.length)))

            : stringIs(varKey, "->list")
              ? quote
                ( fnOfType
                  ( fnLabel
                  , fn =>
                      fnOfType
                      ( intLabel
                      , length =>
                        { if (length < 0) return Null()

                        ; const toReturn = []
                        ; for (let i = 0; i < length; i++)
                            {toReturn.push(fn(makeInt(i)))}
                        ; return makeList(toReturn)})))

            : stringIs(varKey, "true") ? quote(True)

            : stringIs(varKey, "false") ? quote(False)

            : stringIs(varKey, "unit") ? quote(unit)

            : stringIs(varKey, "read-file")
              ? quote
                ( makeFn
                  ( arg =>
                      isString(arg)
                      ? strToChars(readFile(strVal(arg)))
                      : Null('Tried to read-file of nonstring')))

            : stringIs(varKey, "try-parse-prog")
              ? quote
                ( makeFn
                  ( arg =>
                      isString(arg)
                      ? ( parsed =>
                            makeList
                            ( parsed.success
                              ? [True, parsed.ast, strToChars(parsed.rest)]
                              : [ False
                                , makeFn
                                  ( _.flow
                                    (strVal, parsed.error, strToChars))]))
                        (parseFile(strVal(arg)))
                      : Null('Tried to parse nonstring')))

            : stringIs(varKey, "eval-file")
              ? quote
                ( makeFn
                  ( arg =>
                      isString(arg)
                      ? ( parsed =>
                            parsed.success
                            ? topEval(parsed.ast, parsed.rest)
                            : Null(parsed.error(strVal(arg))))
                        (parseFile(readFile(strVal(arg))))
                      : Null('Tried to eval-file of nonstring')))

            : stringIs(varKey, "q") ? quote(makeFn(quote))

            : stringIs(varKey, "make-call")
              ? quote
                ( makeFn
                  (fnExpr => makeFn(argExpr => makeCall(fnExpr, argExpr))))

            : stringIs(varKey, "call-fn-expr")
              ? quote(fnOfType(callLabel, ({fnExpr}) => fnExpr))

            : stringIs(varKey, "call-arg-expr")
              ? quote(fnOfType(callLabel, ({argExpr}) => argExpr))

            : stringIs(varKey, "make-ident") ? quote(makeFn(makeIdent))

            : stringIs(varKey, "ident-key")
              ? quote(fnOfType(identLabel, _.identity))

            : stringIs(varKey, "error")
              ? quote
                ( makeFn
                  ( msgStr =>
                      Null
                      ( isString(msgStr)
                        ? strVal(msgStr)
                        : "Error message wasn't stringy enough")))

            : stringIs(varKey, "bracketed-var") ? quote(bracketedVar)

            : stringIs(varKey, "braced-var") ? quote(bracedVar)

            : stringIs(varKey, "just") ? quote(makeFn(just))

            : stringIs(varKey, "nothing") ? quote(nothing)

            : stringIs(varKey, "is-just")
              ? quote(fnOfType(maybeLabel, arg => makeBool(arg.is)))

            : stringIs(varKey, "unjust")
              ? quote
                ( fnOfType
                  ( maybeLabel
                  , arg => arg.is ? arg.val : Null("Nothing was unjustified")))

            : varKey.data[0].data == '"'.codePointAt(0)
              ? quote(makeList(varKey.data.slice(1)))

            : (varStr => 
                /^(\-|\+)?[0-9]+$/.test(varStr)
                ? quote(makeInt(parseInt(varStr, 10)))
                : Null
                  ( 'string variable not found in environment: "'
                    + strVal(varKey)))
              (strVal(varKey))
        : Null("unknown variable: " + strVal(toString(varKey))))
; exports.initEnv = initEnv

; Error.stackTraceLimit = Infinity

