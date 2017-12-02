'use strict'

// Dependencies

; const
    _ = require('lodash')
  , ps = require('list-parsing')

// Export

; module.exports = parseFile

// Stuff

; function comment()
  { return (
      ps.name
      ( ps.nilStacked
        ( { parseElem
            : elem =>
                ps.parseElem
                ( ps.or
                  ( [ ps.seq([ps.string(';'), ows(), expr()])
                    , ps.seq
                      ( [ ps.string('#')
                        , ps.seq
                          ( [ ps.many(ps.elemNot([ps.string('\u000A')]))
                            , ps.string('\u000A')])])])
                , elem)
          , match: false
          , result: undefined
          , noMore: false
          , futureSuccess: false})
      , "comment"))}

; function ows() {return ps.many(ps.or([ps.wsChar, comment()]))}

; const nameChar
  = ps.elemNot
    ( [ ps.string('(')
      , ps.string(')')
      , ps.string('[')
      , ps.string(']')
      , ps.string('{')
      , ps.string('}')
      , ps.string('`')
      , ps.wsChar
      , ps.string('\\')
      , ps.string('|')
      , ps.string(';')
      , ps.string('#')
      , ps.string('"')])

; function name()
  { return (
      ps.name
      ( ps.map
        ( ps.after(ps.many1(nameChar), ps.wsChar)
        , pt => ['ident', pt.map(chr => chr.codePointAt(0))])
      , "short-identifier"))}

//; const heredoc
//  = ps.name
//    ( ps.map
//      ( ps.then
//        ( ps.around
//          ( ps.string('@')
//          , ps.many
//            ( ps.or
//              ( [ ps.elemNot([ps.string('@'), ps.string('\\')])
//                , ps.before
//                  ( ps.string('\\')
//                  , ps.or([ps.string('@'), ps.string('\\')]))]))
//          , ps.string('@'))
//        , [ endStr =>
//              ps.shortest
//              (ps.after(ps.anything, ps.seq(endStr.map(ps.string))))])
//      , pt => ['heredoc', pt.map(chr => ['char', chr.codePointAt(0)])])
//    , "heredoc")

; const character
  = ps.name
    ( ps.map
      (ps.before(ps.string('`'), ps.oneElem), pt => ['char', pt.codePointAt(0)])
    , "character-literal")

; const ident
  = ps.name
    ( ps.map
      ( ps.around
        ( ps.string('|')
        , ps.many
          ( ps.or
            ( [ ps.elemNot([ps.string('\\'), ps.string('|')])
              , ps.before
                (ps.string('\\'), ps.or([ps.string('|'), ps.string('\\')]))]))
        , ps.string('|'))
      , pt => ['ident', pt.map(chr => chr.codePointAt(0))])
    , "long-identifier")

; function listContents()
  {return ps.around(ows(), ps.sepBy(expr(), ows()), ows())}

//; function parenCall()
//  { return (
//      ps.name
//      ( ps.map
//        ( ps.around(ps.string("("), listContents(), ps.string(")"))
//        , pt => ['call', pt])
//      , "paren-list"))}
//
//; function list()
//  { return (
//      ps.name
//      ( ps.map
//        ( ps.around(ps.string("["), listContents(), ps.string("]"))
//        , pt => ['bracketed', pt])
//      , "bracket-list"))}
//
//; function braced()
//  { return (
//      ps.name
//      ( ps.map
//        ( ps.around(ps.string("{"), listContents(), ps.string("}"))
//        , pt => ['braced', pt])
//      , "brace-list"))}

; function delimited()
  { return (
      ps.name
      ( ps.map
        ( ps.seq
          ( [ ps.or([ps.string("("), ps.string("["), ps.string("{")])
            , listContents()
            , ps.or([ps.string(")"), ps.string("]"), ps.string("}")])])
        , pt => ['delimited', pt])
      , "delimited-list"))}

; function quote()
  { return (
      ps.name
      ( ps.map
        ( ps.before(ps.seq([ps.string('"'), ows()]), expr())
        , pt => ['quote', pt])
      , "quotation"))}

; function expr()
  { return (
      ps.nilStacked
      ( { parseElem
          : elem =>
              ps.parseElem
              ( ps.or
                ( [ delimited()
                  , name()
                  , ident
                  //, heredoc
                  , quote()
                  , character])
              , elem)
        , match: false
        , result: undefined
        , noMore: false
        , futureSuccess: false}))}

; function parseFile(str)
  { const
      arr = _.toArray(str)
    , parsed = ps.shortestMatch(ps.before(ows(), expr()), arr)
  ; return (
      parsed.status === 'match'
      ? _.assign({rest: arr.slice(parsed.index)}, parsed)
      : parsed)}

