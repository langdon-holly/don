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
  , promiseSyncMap
    = (arrIn, promiseFn) =>
        _.reduce
        ( arrIn
        , (prm, nextIn, idx) =>
            prm.then
            ( arrOut =>
                promiseFn(nextIn).then
                (newVal => (arrOut[idx] = newVal, Promise.resolve(arrOut))))
        , Promise.resolve(Array(arrIn.length)))

// Stuff

; exports = module.exports

; function apply(fn, arg)
  { const
      {type: funLabel, data: funData} = fn
    , {type: argLabel, data: argData} = arg
  ; return (
      funLabel === fnLabel
      ? funData(arg)
      : funLabel === listLabel
        ? argLabel !== intLabel
          ? Promise.reject("Argument to list must be integer")
          : argData < 0 || argData >= funData.length
            ? Promise.reject("Array index out of bounds")
            : Promise.resolve(funData[argData])
        : funLabel === quoteLabel
          ? Promise.resolve(funData)
          : funLabel === callLabel
            ? apply(funData.fnExpr, arg).then
              ( fnVal =>
                  apply(funData.argExpr, arg).then
                  (argVal => apply(fnVal, argVal)))
            : funLabel === identLabel
              ? apply(arg, fn).then(expr => apply(expr, arg))
              : funLabel === boolLabel
                ? Promise.resolve
                  ( makeFn
                    ( funData
                    ? _.constant(Promise.resolve(arg))
                    : o => Promise.resolve(o)))
                : Promise.reject("Tried to apply a non-function"))}
; exports.apply = apply

; function Continue(cont, arg)
  {}

; function mk(label, data) {return {type: label, data: data}}

; function makeCall(fnExpr, argExpr) {return mk(callLabel, {fnExpr, argExpr})}

; function fnOfType(type, fn)
  { return (
      makeFn
      ( arg =>
          arg.type !== type
          ? Promise.reject("typed function received garbage")
          : fn(arg.data)))}

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

; function okResult(val) {return mk(resultLabel, {ok: true, val})}

; function errResult(val) {return mk(resultLabel, {ok: false, val})}

; function makeCell(val) {return mk(cellLabel, {val})}

; function makeCont(fn) {return mk(contLabel, {fn})}

; function makeMap(args)
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
        ; return Promise.resolve(toReturn)}))}

; const objToNsNotFoundStr = "Var not found in ns"

; function objToNs(o)
  { return (
      fnOfType
      ( identLabel
      , identKey =>
        { if (!isString(identKey)) return Promise.reject(objToNsNotFoundStr)
        ; const keyStr = strVal(identKey)
        ; return (
            o.hasOwnProperty(keyStr)
            ? Promise.resolve(o[keyStr])
            : Promise.reject
              ( objToNsNotFoundStr
              + ": "
              + strVal(toString(makeIdent(identKey)))))}))}

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
          && eq(val0.data.argExr, val1.data.argExr)
        ||
          val0.type === maybeLabel
          && val0.data.is === val1.data.is
          && (!val0.data.is || eq(val0.data.val, val1.data.val)))
        || val0.type === resultLabel
           && val0.data.ok === val1.data.ok
           && eq(val0.data.val, val1.data.val))}

; function parseTreeToAST(pt)
  { const label = pt[0]
  ; const data = pt[1]

  ; if (label == 'char') return quote(makeChar(data))
  ; if (label == 'call')
      return (
        data.length === 0
        ? quote(makeFn(o => Promise.resolve(o)))
        : _.reduce(_.map(data, parseTreeToAST), makeCall))
  ; if (label == 'bracketed')
      return (
        makeCall
        ( bracketedVar
        , makeFn
          ( env =>
              promiseSyncMap(data.map(parseTreeToAST), expr => apply(expr, env))
              .then(makeList))))
  ; if (label == 'braced')
      return (
        makeCall
        ( bracedVar
        , makeFn
          ( env =>
              promiseSyncMap(data.map(parseTreeToAST), expr => apply(expr, env))
              .then(makeList))))
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

; const resultLabel = {label: 'result'}
; exports.resultLabel = resultLabel

; const cellLabel = {label: 'cell'}
; exports.cellLabel = cellLabel

; const contLabel = {label: 'continuation'}
; exports.contLabel = contLabel

; const bracketedVarSym = gensym('bracketed-var')
; const bracketedVar = makeIdent(bracketedVarSym)
; exports.bracketedVar = bracketedVar

; const bracedVarSym = gensym('braced-var')
; const bracedVar = makeIdent(bracedVarSym)
; exports.bracedVar = bracedVar

; const Null
  = (...args) => {throw new Error("Diverging: " + util.format(...args))}
; exports.Null = Null

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
          return {success: true, ast: parsed.ast, rest: parsed.rest}
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
                    + "^"
                    + util.inspect(parsed.parser.traceStack, {depth: null})})}

        //; const trace = parsed.trace
        //; _.forEachRight(trace, function(frame) {console.log("in", frame[0])})
        //; console.log(parsed.parser)
        }}
; exports.parse = parseFile

; const topEval
  = (ast, rest) =>
      ( promisedQuotedSourceData =>
          apply
          ( ast
          , makeFn
            ( Var =>
                eq(Var, makeIdent(strToChars('source-data')))
                ? promisedQuotedSourceData
                : apply(initEnv, Var))))
      (Promise.resolve(quote(strToChars(rest))))
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

  ; if (argLabel === resultLabel)
      return (
        makeList
        ( strToChars(argData.ok ? "(ok " : "(err ").data.concat
          (toString(argData.val).data, [strToChar(")")])))

  ; if (argLabel === cellLabel)
      return (
        makeList
        ( strToChars("(make-cell ").data.concat
          (toString(argData.val).data, [strToChar(")")])))

  ; if (argLabel === contLabel) return strToChars("(cont ... )")

  ; return Null("->str unknown type:", arg)}

; const initEnv
  = fnOfType
    ( identLabel
    , varKey =>
        varKey.type === symLabel
        ? varKey === bracketedVarSym
          ? Promise.resolve(quote(makeFn(o => Promise.resolve(o))))

          : varKey === bracedVarSym
            ? Promise.resolve
              (quote(fnOfType(listLabel, _.flow(makeMap, o => Promise.resolve(o)))))

            : Promise.reject("symbol variable not found in environment")

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
            ? Promise.resolve
              ( makeFn
                ( env =>
                    Promise.resolve
                    ( makeFn
                      ( param =>
                          Promise.resolve
                          ( makeFn
                            ( body =>
                                Promise.resolve
                                ( makeFn
                                  ( arg =>
                                      apply
                                      ( body
                                      , makeFn
                                        ( varKey =>
                                            eq(varKey, param)
                                            ? Promise.resolve(quote(arg))
                                            : apply(env, varKey)))))))))))

            : stringIs(varKey, '+')
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( listLabel
                    , args =>
                        Promise.resolve().then
                        ( () =>
                            Promise.resolve
                            ( _.reduce
                              ( args
                              , (arg0, arg1) =>
                                {
                                  if (arg1.type !== intLabel)
                                    throw (
                                      strToChars
                                      ("Additional argument wasn't integral"))
                                ; return makeInt(arg0.data + arg1.data)}
                              , makeInt(0)))))))

            : stringIs(varKey, '-')
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( listLabel
                    , args =>
                      { if (args.length === 0)
                          return Promise.resolve(makeInt(-1))

                      ; if (args[0].type !== intLabel) return Promise.reject()
                      ; if (args.length === 1)
                          return Promise.resolve(makeInt(-args[0].data))

                      ; return (
                          Promise.resolve().then
                          ( () =>
                              Promise.resolve
                              ( _.reduce
                                ( args
                                , (arg0, arg1) =>
                                  { if (arg1.type !== intLabel)
                                      throw (
                                        strToChars
                                        ( "Subtractional argument wasn't "
                                          + "integral"))
                                  ; return (
                                      makeInt(arg0.data - arg1.data))}))))})))

            : stringIs(varKey, '<')
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( intLabel
                    , arg0 =>
                        Promise.resolve
                        ( fnOfType
                          ( intLabel
                          , arg1 => Promise.resolve(makeBool(arg0 < arg1)))))))

            : stringIs(varKey, '=')
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( arg0 =>
                        Promise.resolve
                        ( makeFn
                          ( arg1 =>
                              Promise.resolve(makeBool(eq(arg0, arg1))))))))

            : stringIs(varKey, "env")
              ? Promise.resolve(makeFn(o => Promise.resolve(o)))

            : stringIs(varKey, "init-env") ? Promise.resolve(quote(initEnv))

            : stringIs(varKey, "print")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( arg =>
                        isString(arg)
                        ? ( process.stdout.write(strVal(arg))
                          , Promise.resolve(unit))
                        : Promise.reject('Tried to print nonstring'))))

            : stringIs(varKey, "say")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( _.flow
                      ( toString
                      , strVal
                      , process.stdout.write.bind(process.stdout)
                      , _.constant(Promise.resolve(unit))))))

            : stringIs(varKey, "->str")
              ? Promise.resolve
                (quote(makeFn(_.flow(toString, o => Promise.resolve(o)))))

            : stringIs(varKey, "char->unicode")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (charLabel, _.flow(makeInt, o => Promise.resolve(o)))))

            : stringIs(varKey, "unicode->char")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (intLabel, _.flow(makeChar, o => Promise.resolve(o)))))

            : stringIs(varKey, "length")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (listLabel, arg => Promise.resolve(makeInt(arg.length)))))

            : stringIs(varKey, "->list")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( fnLabel
                    , fn =>
                        Promise.resolve
                        ( fnOfType
                          ( intLabel
                          , length =>
                            { if (length < 0) return Promise.reject()

                            ; const toReturn = []
                            ; for (let i = 0; i < length; i++)
                                toReturn.push(fn(makeInt(i)))
                            ; return Promise.resolve(makeList(toReturn))})))))

            : stringIs(varKey, "true") ? Promise.resolve(quote(makeBool(true)))

            : stringIs(varKey, "false")
              ? Promise.resolve(quote(makeBool(false)))

            : stringIs(varKey, "unit") ? Promise.resolve(quote(unit))

            : stringIs(varKey, "read-file")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( arg =>
                        isString(arg)
                        ? Promise.resolve(strToChars(readFile(strVal(arg))))
                        : Promise.reject('Tried to read-file of nonstring'))))

            : stringIs(varKey, "parse-prog")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( arg =>
                        isString(arg)
                        ? Promise.resolve
                          ( ( parsed =>
                                parsed.success
                                ? okResult
                                  ( objToNs
                                    ( { "expr": quote(parsed.ast)
                                      , "rest": quote(strToChars(parsed.rest))}))
                                : errResult
                                  ( makeFn
                                    ( _.flow
                                      (strVal, parsed.error, strToChars))))
                            (parseFile(strVal(arg))))
                        : Promise.reject('Tried to parse nonstring'))))

            : stringIs(varKey, "eval-file")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( arg =>
                        isString(arg)
                        ? ( parsed =>
                              parsed.success
                              ? topEval(parsed.ast, parsed.rest)
                              : Promise.reject(parsed.error(strVal(arg))))
                          (parseFile(readFile(strVal(arg))))
                        : Promise.reject('Tried to eval-file of nonstring'))))

            : stringIs(varKey, "q")
              ? Promise.resolve
                (quote(makeFn(_.flow(quote, o => Promise.resolve(o)))))

            : stringIs(varKey, "make-call")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( fnExpr =>
                        Promise.resolve
                        ( makeFn
                          ( argExpr =>
                              Promise.resolve(makeCall(fnExpr, argExpr)))))))

            : stringIs(varKey, "call-fn-expr")
              ? Promise.resolve
                ( quote
                  (fnOfType(callLabel, ({fnExpr}) => Promise.resolve(fnExpr))))

            : stringIs(varKey, "call-arg-expr")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (callLabel, ({argExpr}) => Promise.resolve(argExpr))))

            : stringIs(varKey, "make-ident")
              ? Promise.resolve
                (quote(makeFn(_.flow(makeIdent, o => Promise.resolve(o)))))

            : stringIs(varKey, "ident-key")
              ? Promise.resolve
                ( quote
                  (fnOfType(identLabel, o => Promise.resolve(o))))

            : stringIs(varKey, "error")
              ? Promise.resolve
                ( quote
                  ( makeFn
                    ( msgStr =>
                        Promise.reject
                        ( isString(msgStr)
                          ? strVal(msgStr)
                          : "Error message wasn't stringy enough"))))

            : stringIs(varKey, "bracketed-var")
              ? Promise.resolve(quote(bracketedVar))

            : stringIs(varKey, "braced-var") ? Promise.resolve(quote(bracedVar))

            : stringIs(varKey, "just")
              ? Promise.resolve
                (quote(makeFn(_.flow(just, o => Promise.resolve(o)))))

            : stringIs(varKey, "nothing") ? Promise.resolve(quote(nothing))

            : stringIs(varKey, "justp")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (maybeLabel, arg => Promise.resolve(makeBool(arg.is)))))

            : stringIs(varKey, "unjust")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( maybeLabel
                    , arg =>
                        arg.is
                        ? Promise.resolve(arg.val)
                        : Promise.reject("Nothing was unjustified"))))

            : stringIs(varKey, "ok")
              ? Promise.resolve
                (quote(makeFn(_.flow(okResult, o => Promise.resolve(o)))))

            : stringIs(varKey, "err")
              ? Promise.resolve
                (quote(makeFn(_.flow(errResult, o => Promise.resolve(o)))))

            : stringIs(varKey, "okp")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    (resultLabel, arg => Promise.resolve(makeBool(arg.ok)))))

            : stringIs(varKey, "unok")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( resultLabel
                    , arg =>
                        arg.ok
                        ? Promise.resolve(arg.val)
                        : Promise.reject
                          ( isString(arg.val)
                            ? "Err: " + strVal(arg.val)
                            : "Result was not ok"))))

            : stringIs(varKey, "unerr")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( resultLabel
                    , arg =>
                        arg.ok
                        ? Promise.reject
                          ( isString(arg.val)
                            ? "Ok: " + strVal(arg.val)
                            : "Result was ok")
                        : Promise.resolve(arg.val))))

            : stringIs(varKey, "make-cell")
              ? Promise.resolve
                (quote(makeFn(_.flow(makeCell, o => Promise.resolve(o)))))

            : stringIs(varKey, "cell-val")
              ? Promise.resolve
                (quote(fnOfType(cellLabel, ({val}) => Promise.resolve(val))))

            : stringIs(varKey, "set")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( cellLabel
                    , cell =>
                        Promise.resolve
                        ( makeFn
                          (val => (cell.val = val, Promise.resolve(unit)))))))

            : stringIs(varKey, "cas")
              ? Promise.resolve
                ( quote
                  ( fnOfType
                    ( cellLabel
                    , cell =>
                        Promise.resolve
                        ( makeFn
                          ( oldVal =>
                              Promise.resolve
                              ( makeFn
                                ( newVal =>
                                    Promise.resolve
                                    ( eq(cell.val, oldVal)
                                      ? (cell.val = newVal, oldVal)
                                      : newVal))))))))

            : varKey.data[0].data == '"'.codePointAt(0)
              ? Promise.resolve(quote(makeList(varKey.data.slice(1))))

            : (varStr =>
                /^(\-|\+)?[0-9]+$/.test(varStr)
                ? Promise.resolve(quote(makeInt(parseInt(varStr, 10))))
                : Promise.reject
                  ( 'string variable not found in environment: "'
                    + strVal(varKey)))
              (strVal(varKey))
        : Promise.reject("unknown variable: " + strVal(toString(varKey))))
; exports.initEnv = initEnv

; Error.stackTraceLimit = Infinity

