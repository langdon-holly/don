require('lodash');

function expr(str) {
  if (str.charAt(0) === '(') {
    return form(str);
  }
}

function form(str) {
  if (str.charAt(0) === '('
      && str.charAt(str.length-1) === ')') {
    
  }
  return [false]
}

function seq() {
  var args = arguments;
  if (args.length < 2) return [false];
  if (args.length > 2) {
    
  }
}
