"use strict"

var bigInt = require('big-integer')

; module.exports
  = { fromNumber: function(num) {return bigInt(num)}
    , zero: bigInt.zero
    , isZero: function(nat) {return nat.isZero()}
    , succ: function(nat) {return nat.next()}
    , pred: function(nat) {return nat.isZero() ? [] : [nat.prev()]}
    , eq:   function(nat0, nat1) {return nat0.equals(nat1)}
    , add:  function(nat0, nat1) {return nat0.add(nat1)}
    , sub
      :  function(minuend, subtrahend) {return minuend.subtract(subtrahend)}
    , cmp:  function(nat0, nat1) {return nat0.compare(nat1)}
    , div
      : function(numerator, denominator)
          {return numerator.divide(denominator)}
    , mod
      : function(numerator, denominator) {return numerator.mod(denominator)}
    , divMod
      : function(numerator, denominator)
          {return numerator.divmod(denominator)}
    , mul: function(nat0, nat1) {return nat0.multiply(nat1)}}

