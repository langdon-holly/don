var _ = require('lodash');
var ps = require('list-parsing');

module.exports = parseFile;

function comment() {
  return {parseElem: function(chr) {
            return ps.or(ps.seq(ps.string(';'),
                                ows(),
                                expr),
                         ps.seq(ps.string('#'),
                                ps.seq(ps.many(ps.charNot(ps.string('\u000A'))),
                                       ps.string('\u000A')))).parseElem(chr);},
          result: [false],
          noMore: false,
          futureSuccess: false};}

function oComment() {return ps.many(comment());}

function ows() {return ps.many(ps.or(ps.wsChar, comment()));}

function theWs() {return ps.seq(ps.many(comment()), ps.wsChar, ows());}

function call
() {
  return ps
         .mapParser
         ( ps.before
           ( ps.seq(ps.string('\\'), ows())
           , ps.and(ps.between(ps.nothing, expr, ows(), expr), ps.not(expr)))
         , function(pt) {return ['call', pt[0]];});}

var nameChar = ps.charNot(ps.string('('),
                          ps.string(')'),
                          ps.string('['),
                          ps.string(']'),
                          ps.string('{'),
                          ps.string('}'),
                          ps.string('`'),
                          ps.wsChar,
                          ps.string('\\'),
                          ps.string('|'),
                          ps.string(';'),
                          ps.string('#'))

var nameBegin = ps.seq(nameChar, ps.anything);

function name() {
  return ps.mapParser
         ( ps.after(ps.many1(nameChar), ps.opt(ps.wsChar)),
           function(pt)
           {return [ 'call'
                   , pt.map
                     (function(chr) {return ['char', chr.codePointAt(0)];})];});}

var heredoc
= ps.mapParser(
    ps.then(
      ps.around(
        ps.string('`'),
        ps.many(ps.or(ps.charNot(ps.string('`'), ps.string('\\')),
                      ps.before(ps.string('\\'),
                                ps.or(ps.string('`'), ps.string('\\'))))),
        ps.string('`')),
      function(endStr) {
        return ps.shortest(
                 ps.after(
                   ps.anything,
                   ps.seq.apply(
                     this,
                     endStr.map(function(chr) {return ps.string(chr);}))));}),
    function(pt) {return ['heredoc', pt.split('').map(function(chr) {return ['char', chr.codePointAt(0)]})];});

var braceStr = {
  parseElem: function(chr) {
    return ps.mapParser
           ( ps.around
             ( ps.string('{'),
               ps.many
               ( ps.or
                 ( ps.mapParser(
                     ps.charNot(ps.string('|')),
                     function(pt) {
                       return ['str', [pt]];}),
                   ps.before(
                     ps.string('|'),
                     ps.or(
                       ps.mapParser(
                         ps.string('|'),
                         function() {return ['str', ['|']];}),
                       ps.before(
                         ows(),
                         ps.mapParser(
                           expr
                           , function(pt) {return ['expr', pt];})))))),
               ps.string('|}')),
             function(arr) {
               var toReturn = [];
               _
               .forEach
               ( arr,
                 function(elem) {
                   if (elem[0] === 'comment') return;
                   if
                     ( toReturn.length == 0
                       || elem[0] === 'expr')
                     {toReturn.push(elem); return;}
                   var last = _.last(toReturn);
                   if
                     (last[0] === 'str')
                     {last[1].push(elem[1][0]); return;}
                   toReturn.push(elem);});

               return ['braceStr',
                       toReturn];}).parseElem(chr);},
  result: [false],
  noMore: false,
  futureSuccess: false}

function
  listContents
  ()
  {return ps.around(ows(), ps.sepBy(expr, theWs()), ows());}

function parenCall() {
  return ps.mapParser(ps.around(ps.string("("),
                                listContents(),
                                ps.string(")")),
                      function(pt) {
                        return ['call', pt];});}

function list() {
  return ps.mapParser(ps.around(ps.string("["),
                                listContents(),
                                ps.string("]")),
                      function(pt) {
                        return ['list', pt];});}

var expr = {parseElem: function(chr) {
              return ps.or(parenCall(),
                           list(),
                           name(),
                           call(),
                           braceStr,
                           heredoc).parseElem(chr);},
            result: [false],
            noMore: false,
            futureSuccess: false};

function parseFile(str) {
  var parsed = ps.longestMatch(ps.before(ows(), expr), str);
  if (parsed[0][0]) {
    return [parsed[0], str.slice(parsed[1], str.length)];}
  return [[false, parsed[1]]];}

