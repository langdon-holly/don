var _ = require('lodash');
var ps = require('list-parsing');

module.exports = parseFile;

function comment() {
  return {parseChar: function(chr) {
            return ps.or(ps.seq(ps.string(';'),
                                ows(),
                                exprs()),
                         ps.seq(ps.string('#'),
                                ps.seq(ps.many(ps.charNot(ps.string('\u000A'))),
                                       ps.string('\u000A'))),
                         ps.seq(ps.string('\\'), ows())).parseChar(chr);},
          result: [false],
          noMore: false,
          futureSuccess: false};}

function oComment() {return ps.many(comment());}

function ows() {return ps.many(ps.or(ps.wsChar, comment()));}

function theWs() {return ps.seq(ows(), ps.wsChar, ows());}

function exprs() {
  return {parseChar: function(chr) {
            return ps.mapParser(
                        ps.or(ps.mapParser(expr, function(pt) {return [pt];}),
                              ps.between(ps.nothing,
                                         ps.mapParser(name(),
                                                      function(pt) {
                                                        return [pt];}),
                                         oComment(),
                                         ps.and(exprs(), ps.not(nameBegin))),
                              ps.between(ps.nothing,
                                         ps.and(expr, ps.not(nameBegin)),
                                         oComment(),
                                         ps.and(exprs(), nameBegin)),
                              ps.between(ps.nothing,
                                         ps.and(expr, ps.not(nameBegin)),
                                         oComment(),
                                         ps.and(exprs(), ps.not(nameBegin)))),
                        function(pt) {
                          if (pt.length == 1) return pt[0];
                          return ['form', 
                                  [pt[0][0],
                                   pt[1][0]]];}).parseChar(chr);},
          result: [false],
          noMore: false,
          futureSuccess: false};}

var nameChar = ps.charNot(ps.string('('),
                          ps.string(')'),
                          ps.string('['),
                          ps.string(']'),
                          ps.string('{'),
                          ps.string('}'),
                          ps.string('`'),
                          ps.wsChar,
                          ps.string('\\'),
                          ps.string(';'),
                          ps.string('#'))

var nameBegin = ps.seq(nameChar, ps.anything);

function name() {
  return ps.mapParser(ps.mapParser(ps.many1(nameChar),
                                   function concatStrs(arr) {
                                     if (arr.length == 1) return arr[0];
                                     if (arr.length == 2) return arr[0]
                                       .concat(arr[1]);
                                     return arr[0].concat(
                                       concatStrs(arr.slice(1,
                                                            arr.length)))}),
                      function(pt) {return ['name', pt];});}

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
    function(pt) {return ['heredoc', pt];});

var braceStr = {
  parseChar: function(chr) {
    return ps.mapParser(
             ps.around(
               ps.string('{'),
               ps.many(ps.or(ps.mapParser(ps.charNot(ps.string('{'),
                                                     ps.string('}'),
                                                     ps.string('\\'),
                                                     ps.string('`')),
                                          function(pt) {
                                            return ['str', pt];}),
                             ps.mapParser(ps.before(ps.string('\\'),
                                                    ps.charNot()),
                                          function(pt) {
                                            if (pt === '\\'
                                                || pt === '{'
                                                || pt === '}'
                                                || pt === '`')
                                              return ['str', pt];
                                            return ['escape', pt];}),
                             ps.mapParser(ps.around(ps.string('{'),
                                                    listContents(),
                                                    ps.string('}')),
                                          function(pt) {return ['form', pt];}),
                             ps.mapParser(heredoc,
                                          function(pt) {
                                            return ['str', pt[1]];}))),
               ps.string('}')),
             function(arr) {
               var toReturn = [];
               _
               .forEach
               ( arr,
                 function(elem) {
                   if
                     ( toReturn.length == 0
                       || elem[0] === 'form'
                       || elem[0] === 'escape')
                     {toReturn.push(elem); return;}
                   var last = _.last(toReturn);
                   if
                     (last[0] === 'str' || last[0] === 'escape')
                     {last[1] = last[1] + elem[1]; return;}
                   toReturn.push(elem);});

               return ['braceStr',
                       toReturn];}).parseChar(chr);},
  result: [false],
  noMore: false,
  futureSuccess: false}

function
  listContents
  ()
  {return ps.around(ows(), ps.sepBy(exprs(), theWs()), ows());}

function form() {
  return ps.mapParser(ps.around(ps.string("("),
                                listContents(),
                                ps.string(")")),
                      function(pt) {
                        return ['form', pt];});}

function list() {
  return ps.mapParser(ps.around(ps.seq(ps.string("["),
                                       ows()),
                                ps.sepBy(exprs(), theWs()),
                                ps.seq(ows(),
                                       ps.string("]"))),
                      function(pt) {
                        return ['list', pt];});}

var expr = {parseChar: function(chr) {
              return ps.or(form(),
                           list(),
                           name(),
                           braceStr,
                           heredoc).parseChar(chr);},
            result: [false],
            noMore: false,
            futureSuccess: false};

function parseFile(str) {
  var parsed = ps.longestMatch(ps.before(ows(), exprs()), str);
  if (parsed[0][0]) {
    return [parsed[0], str.slice(parsed[1], str.length)];}
  return [[false, parsed[1]]];}

