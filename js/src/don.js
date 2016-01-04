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
    for (var betweenI; betweenI >= 0; betweenI--) {
      
    }
  }
  else if (args.length > 2) {
    return seq(args[0], seq.apply(this, Array.prototype.slice.call(args, 1, args.length)));
  }
}
