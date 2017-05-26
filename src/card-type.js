'use strict'

// Dependencies

; var
    _   = require('lodash')
  , nat = require('./nat.js')
  , two = nat.fromNumber(2)

// Stuff

; module.exports
  = { makeProduct
      : function(nat0, card0, nat1, card1)
        { if (!nat.isZero(card1)) return nat.add(nat.mul(nat0, card1), nat1)
        ; if (!nat.isZero(card0)) return nat.add(nat.mul(nat1, card0), nat0)
        ; var sum = nat.add(nat0, nat1)
        ; return nat.add(nat.div(nat.mul(nat.succ(sum), sum), two), nat1)}}

