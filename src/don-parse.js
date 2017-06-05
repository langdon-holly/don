'use strict'

// Dependencies

; var
    _ = require('lodash')
  , ps = require('list-parsing')

// Export

; module.exports = parseFile

// Stuff

; function comment()
  {return (
     ps.name
     ( { parseElem
         : function(elem)
           {return (
              ps.or
              ( ps.seq(ps.string(';'), ows(), expr)
              , ps.seq
                ( ps.string('#')
                , ps.seq
                  ( ps.many(ps.elemNot(ps.string('\u000A')))
                  , ps.string('\u000A')))).parseElem(elem))}
       , match: false
       , result: undefined
       , noMore: false
       , futureSuccess: false
       , stack: ps.nil}
     , "comment"))}

; function ows() {return ps.many(ps.or(ps.wsChar, comment()))}

; function call()
  {return (
     ps.name
     ( ps.map
       ( ps.before
         ( ps.seq(ps.string('\\'), ows())
         , ps.between(ps.nothing, expr, ows(), expr))
       , function(pt) {return ['call', pt]})
     , "function call"))}

; var nameChar
  = ps.elemNot
    ( ps.string('(')
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
    , ps.string('"'))

; function name()
  {return (
     ps.name
     ( ps.map
       ( ps.after(ps.many1(nameChar), ps.wsChar)
       , function(pt)
         {return ['ident', pt.map(function(chr) {return chr.codePointAt(0)})]})
     , "short-identifier"))}

; var heredoc
  = ps.name
    ( ps.map
      ( ps.then
        ( ps.around
          ( ps.string('`')
          , ps.many
            ( ps.or
              ( ps.elemNot(ps.string('`'), ps.string('\\'))
              , ps.before
                ( ps.string('\\')
                , ps.or(ps.string('`'), ps.string('\\')))))
          , ps.string('`'))
        , function(endStr)
          {return (
             ps.shortest
             ( ps.after
               ( ps.anything
               , ps.seq.apply
                 ( this
                 , endStr.map(function(chr) {return ps.string(chr)})))))})
      , function(pt)
        {return (
           [ 'heredoc'
           , pt.map(function(chr) {return ['char', chr.codePointAt(0)]})])})
    , "heredoc")

; var ident
  = ps.name
    ( ps.map
      ( ps.around
        ( ps.string('|')
        , ps.many
          ( ps.or
            ( ps.elemNot(ps.string('\\'), ps.string('|'))
            , ps.before
              (ps.string('\\'), ps.or(ps.string('|'), ps.string('\\')))))
        , ps.string('|'))
      , function(pt)
        {return ['ident', pt.map(function(chr) {return chr.codePointAt(0)})]})
    , "long-identifier")

; function listContents()
  {return ps.around(ows(), ps.sepBy(expr, ows()), ows())}

; function parenCall()
  {return (
     ps.name
     ( ps.map
       ( ps.around(ps.string("("), listContents(), ps.string(")"))
       , function(pt) {return ['call', pt]})
     , "paren-list"))}

; function list()
  {return (
     ps.name
     ( ps.map
       ( ps.around(ps.string("["), listContents(), ps.string("]"))
       , function(pt) {return ['bracketed', pt]})
     , "bracket-list"))}

; function braced()
  {return (
     ps.name
     ( ps.map
       ( ps.around(ps.string("{"), listContents(), ps.string("}"))
       , function(pt) {return ['braced', pt]})
     , "brace-list"))}

; function quote()
  {return (
     ps.name
     ( ps.map
       ( ps.before
         ( ps.seq(ps.string('"'), ows())
         , expr)
       , function(pt) {return ['quote', pt]})
     , "quotation"))}

; var expr
  = { parseElem
      : function(elem)
        { return (
            ps.or
            ( parenCall()
            , list()
            , name()
            , call()
            , braced()
            , ident
            , heredoc
            , quote())
            .parseElem(elem))}
    , match: false
    , result: undefined
    , noMore: false
    , futureSuccess: false
    , stack: ps.nil}

; function parseFile(str)
  { var arr = _.toArray(str)
  ; var parsed = ps.shortestMatch(ps.before(ows(), expr), arr)
  ; return (
      parsed.status === 'match'
      ? _.assign({rest: arr.slice(parsed.index)}, parsed)
      : parsed)}

