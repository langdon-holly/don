if (!Array.prototype.includes) {
  Array.prototype.includes = function(searchElement /*, fromIndex*/ ) {
    'use strict';
    var O = Object(this);
    var len = parseInt(O.length) || 0;
    if (len === 0) {
      return false;
    }
    var n = parseInt(arguments[1]) || 0;
    var k;
    if (n >= 0) {
      k = n;
    } else {
      k = len + n;
      if (k < 0) {k = 0;}
    }
    var currentElement;
    while (k < len) {
      currentElement = O[k];
      if (searchElement === currentElement ||
         (searchElement !== searchElement && currentElement !== currentElement)) { // NaN !== NaN
        return true;
      }
      k++;
    }
    return false;
  };
}

function valObj(label, data) {
  return [true, [[label, data]]];}

function getInterfaceData(o, targetLabel, implementszes) {
  if (!o[0]) return [false];

  for (var i = 0; i < o[i].length; i++) {
    var maybeInterfaceData = valInterfaceData(o[1][i],
                                              targetLabel,
                                              implementszes);
    if (maybeInterfaceData[0]) return maybeInterfaceData[1];} 
  return [false];}

function valInterfaceData(val, targetLabel, implementszes, prevInterfaces) {
  if (prevInterfaces === undefined) prevInterfaces = [];

  if (prevInterfaces.includes(val[0])) return [false];

  if (val[0] === targetLabel) return [true, val[1]];

  prevInterfaces.push(val[0]);

  var applicableImplementszes = implementszes[val[0]];
  for (var i = 0; i < applicableImplementszes.length; i++) {
    var maybeVal = applicableImplementszes[i](val[1]);
    if (!maybeVal[0]) continue;
    var maybeInterfaceData = valInterfaceData(maybeVal[1],
                                              targetLabel,
                                              implementszes,
                                              prevInterfaces);
    if (maybeInterfaceData[0]) return maybeInterfaceData[1];}
  return [false];}

function mApply(macro, args, implementszes, env) {
  var transformMaybe = getInterfaceData(macro, macroLabel, implementszes);
  if (!transformMaybe[0]) return [false];

  return transformMaybe[1](args, implementszes, env);}

function apply(fn, args, implementszes, env) {
  var transformMaybe = getInterfaceData(fn, fnLabel, implementszes);
  if (!transformMaybe[0]) return [false];

  return transformMaybe[1](args, implementszes, env);}

function eval(form, implementszes, env) {
  var listMaybe = getInterfaceData();}

function fnOfTypes(interfaceList, fn) {
  }

var macroLabel = {};
var fnLabel = {};
var listLabel = {};
var numLabel = {};

var initImplementszes = {};
initImplementszes[listLabel] = [function(list) {
                                  return [true,
                                          valObj(fnLabel,
                                                 function(index) {
                                                   })];}];

