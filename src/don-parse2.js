var _ = require('lodash');
var ps = require('./parse2.js');

module.exports = parseFile;

function comment() {
  return ps.or(ps.seq(ps.string(';'),
                      exprs()),
               ps.seq(ps.string('#'),
                      ps.seq(ps.many(ps.charNot(ps.string('\u000A'))),
                             ps.string('\u000A'))));}

function oComment() {return ps.many(comment());}

function ows() {return ps.many(ps.or(ps.wsChar, comment()));}

function theWs() {return ps.seq(ows(), ps.wsChar, ows());}

function exprs() {
  return ps.mapParser(ps.many1(expr),
                      function reduceExprs(pt) {
                        if (pt.length == 1) return pt[0];
                        return ['form', 
                                [pt[0],
                                 reduceExprs(pt.slice(1, pt.length))]];});}

function name() {
  return ps.mapParser(ps.mapParser(ps.many1(ps.charNot(ps.string('('),
                                                       ps.string(')'),
                                                       ps.string('['),
                                                       ps.string(']'),
                                                       ps.string('{'),
                                                       ps.string('}'),
                                                       ps.wsChar,
                                                       ps.string('\\'),
                                                       ps.string(';'),
                                                       ps.string('#'))),
                                   function concatStrs(arr) {
                                     if (arr.length == 1) return arr[0];
                                     if (arr.length == 2) return arr[0]
                                       .concat(arr[1]);
                                     return arr[0].concat(
                                       concatStrs(arr.slice(1,
                                                            arr.length)))}),
                      function(pt) {return ['name', pt];});}

var braceStr = {
  parseChar: function(chr) {
    return ps.mapParser(
             ps.around(
               ps.string('{'),
               ps.many(ps.or(ps.mapParser(ps.before(ps.string('\\'),
                                                    ps.charNot()),
                                          function(pt) {
                                            if (pt === '\\'
                                                || pt === '{'
                                                || pt === '}') return [pt];
                                            return ['', pt];}),
                             ps.mapParser(braceStr,
                                          function(pt) {
                                            var arr0 = pt[1]
                                            var arr1 = ['{'.concat(arr0[0])]
                                              .concat(arr0.slice(1,
                                                                 arr0.length));
                                            return arr1.slice(0, arr1.length-1)
                                              .concat([arr1[arr1.length-1]
                                                .concat('}')]);}),
                             ps.mapParser(ps.charNot(ps.string('{'),
                                                     ps.string('}'),
                                                     ps.string('\\')),
                                          function(pt) {
                                            return [pt];}))),
               ps.string('}')),
             function(arr) {
               return ['braceStr',
                       _.reduce(arr,
                                function (arr0, arr1) {
                                  return arr0.slice(0, arr0.length - 1)
                                    .concat([arr0[arr0.length - 1]
                                      .concat(arr1[0])])
                                    .concat(arr1.slice(1, arr1.length))},
                                [''])];}).parseChar(chr);},
  result: [false]}

function form() {
  return ps.mapParser(ps.around(ps.seq(ps.string("("),
                                       ows()),
                                ps.sepBy(exprs(), theWs()),
                                ps.seq(ows(),
                                       ps.string(")"))),
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
                           braceStr).parseChar(chr);},
            result: [false],
            doomed: false};

function parseFile(str) {
  var parsed = ps.longestMatch(ps.before(ows(), exprs()), str);
  if (parsed[0][0]) {
    return [parsed[0], str.slice(parsed[1], str.length)];}
  return [[false]];}

