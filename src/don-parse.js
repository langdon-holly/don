var _ = require('lodash');
var ps = require('./parse.js');

module.exports = parseFile;

function expr(str) {
  return ps.or(
    form,
    list,
    name,
    ps.mapParser(bracketStr,
                 function(pt) {
                   if (pt[1].length == 1) return ['name', pt[1][0]];
                   return pt;}))(str);}

function form(str) {
  return ps.mapParser(ps.around(ps.seq(ps.string("("),
                                       ows),
                                ps.sepBy(exprs, theWs),
                                ps.seq(ows,
                                       ps.string(")"))),
                      function(pt) {
                        return ['form', pt];})(str);}

function list(str) {
  return ps.mapParser(ps.around(ps.seq(ps.string("["),
                                       ows),
                                ps.sepBy(exprs, theWs),
                                ps.seq(ows,
                                       ps.string("]"))),
                   function(pt) {
                     return ['list', pt];})(str);}

function theWs(str) {
  return ps.seq(ows, ps.wsChar, ows)(str);}

function ows(str) {
  return ps.many(ps.or(ps.wsChar, comment))(str);}

function exprs(str) {
  return ps.sepBy(expr, oComment)(str);}

function comment(str) {
  return ps.seq(ps.string(';'),
                ps.or(ps.seq(ps.wsChar,
                             ps.many(ps.charNot(ps.string('\u000A'))),
                             ps.string('\u000A')),
                      expr))(str);}

function oComment(str) {
  return ps.many(comment)(str);}

function name(str) {
  return ps.mapParser(ps.mapParser(ps.many1(ps.charNot(ps.string('('),
                                                       ps.string(')'),
                                                       ps.string('['),
                                                       ps.string(']'),
                                                       ps.string('{'),
                                                       ps.string('}'),
                                                       ps.wsChar,
                                                       ps.string('\\'),
                                                       ps.string(';'))),
                                   function concatStrs(arr) {
                                     if (arr.length == 1) return arr[0];
                                     if (arr.length == 2) return arr[0]
                                       .concat(arr[1]);
                                     return arr[0].concat(
                                       concatStrs(arr.slice(1, arr.length)))}),
                      function(pt) {return ['name', pt];})(str);}

function bracketStr(str) {
return ps.mapParser(
  ps.around(ps.string('{'),
            ps.many(ps.or(ps.mapParser(ps.before(ps.string('\\'),
                                                 ps.charNot()),
                                       function(pt) {
                                         if (pt === '\\'
                                             || pt === '{'
                                             || pt === '}') return [pt];
                                         return ['', pt];}),
                          ps.mapParser(bracketStr,
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
    return ['bracketStr',
            _.reduce(arr,
                     function (arr0, arr1) {
                       return arr0.slice(0, arr0.length-1)
                         .concat([arr0[arr0.length-1]
                           .concat(arr1[0])])
                         .concat(arr1.slice(1, arr1.length))},
                     [''])];})(str);}

function parseFile(str) {
  var parsed = ps.longestMatch(expr, str);
  if (parsed[0][0]) {
    return [parsed[0], str.slice(parsed[1], str.length)];}
  return [[false]];}

//console.log(parseFile('(+ {2} 2) hello!'));

