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

  if (args.length == 1) return args[0];
  else if (args.length == 2) {
    
  }
  else if (args.length > 2) {
    return seq(Array.prototype.slice.call(args, 0, args.length-1), args[args.length-1]);
  }
}
