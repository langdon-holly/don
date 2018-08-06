'use strict';

// Dependencies

const _ = require('lodash'), ps = require('list-parsing');

// Export

module.exports = {parseStream, parseIter};

// Stuff

const
  [ lParen
  , rParen
  , lBracket
  , rBracket
  , lBrace
  , rBrace
  , backtick
  , backslash
  , pipe
  , semicolon
  , hash
  , bang
  , dQuote
  , space
  , tab
  , cr
  , lf]
  = Array.from('()[]{}`\\|;#!" \t\r\n').map(ps.string)
  , wsChar = ps.name(ps.or([space, lf, tab, cr]), "ws-char");

const shebang = ps.seq([hash, bang, ps.many(ps.elemNot([lf])), lf]);

function comment()
{ return (
    ps.name
    ( ps.nilStacked
      ( { parseElem
          : elem => ps.parseElem(ps.seq([semicolon, ows(), expr()]), elem)
        , match: false
        , result: undefined
        , noMore: false
        , futureSuccess: false})
    , "comment"))}

function ows() {return ps.many(ps.or([wsChar, comment()]))}

const
  name
  = ps.name
    ( ps.map
      ( ps.after
        ( ps.many1
          ( ps.elemNot
            ( [ wsChar
              , lParen
              , rParen
              , lBracket
              , rBracket
              , lBrace
              , rBrace
              , backtick
              , backslash
              , pipe
              , semicolon
              , hash
              , dQuote]))
        , wsChar)
      , pt => ['ident', pt.map(chr => chr.codePointAt(0))])
    , "short-identifier");

const
  character
  = ps.name
    ( ps.map
      (ps.before(backtick, ps.oneElem), pt => ['char', pt.codePointAt(0)])
    , "character-literal");

const
  identContents
  = nested =>
    ps.nilStacked
    ( { parseElem
        : elem =>
          ps.parseElem
          ( ps.map
            ( ps.around
              ( ps.seq([backslash, ps.many(comment())])
              , ps.sepBy
                ( ps.or
                  ( [ ps.map
                      ( ps.elemNot([backslash, pipe, backtick, semicolon])
                      , Array.of)
                    , nested
                      ? ps.seq([backtick, ps.oneElem])
                      : ps.map(ps.before(backtick, ps.oneElem), Array.of)
                    , identContents(true)])
                , ps.many(comment()))
              , ps.seq([ps.many(comment()), pipe]))
            , _.flow(nested ? a => [['\\'], ...a, ['|']] : o=>o, _.flatten))
          , elem)
      , match: false
      , result: undefined
      , noMore: false
      , futureSuccess: false});

const
  ident
  = ps.name
    ( ps.map
      (identContents(), pt => ['ident', pt.map(chr => chr.codePointAt(0))])
    , "long-identifier");

function listContents()
{return ps.around(ows(), ps.sepBy(expr(), ows()), ows())}

function delimited()
{ return (
    ps.name
    ( ps.map
      ( ps.seq
        ( [ ps.or([lParen, lBracket, lBrace])
          , listContents()
          , ps.or([rParen, rBracket, rBrace])])
      , pt => ['delimited', pt])
    , "delimited-list"))}

function Delimited()
{ return (
    ps.nilStacked
    ( { parseElem
        : elem =>
          ps.parseElem
          ( ps.name
            ( ps.map
              ( ps.seq
                ( [ ps.or([lParen, lBracket, lBrace, backslash])
                  , ps.many
                    ( ps.or
                      ( [ character
                        , Delimited()
                        , ps.map
                          ( ps.elemNot
                            ( [ lParen
                              , lBracket
                              , lBrace
                              , backslash
                              , rParen
                              , rBracket
                              , rBrace
                              , pipe
                              , backtick])
                          , pt => ['elem', pt.codePointAt(0)])]))
                  , ps.or([rParen, rBracket, rBrace, pipe])])
              , pt => ['Delimited', pt])
            , "delimited")
          , elem)
    , match: false
    , result: undefined
    , noMore: false
    , futureSuccess: false}));}

function quote()
{ return (
    ps.name
    ( ps.map(ps.before(ps.seq([dQuote, ows()]), expr()), pt => ['quote', pt])
    , "quotation"))}

function expr()
{ return (
    ps.nilStacked
    ( { parseElem
        : elem =>
          ps.parseElem
          (ps.or([Delimited(), /*delimited(), */name, /*ident, */quote(), character]), elem)
      , match: false
      , result: undefined
      , noMore: false
      , futureSuccess: false}))}

const fileParser = ps.before(ps.seq([ps.opt(shebang), ows()]), expr());

function parseStream(str) {return ps.streamShortestMatch(fileParser, str)}

function parseIter(asyncIterable)
{return ps.asyncIterableShortestMatch(fileParser, asyncIterable)}

