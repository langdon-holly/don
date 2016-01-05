require('lodash');

function expr(str) {
  if (str.charAt(0) === '(') {
    return form(str);
  }
}

function form(str) {
  return seq(string('('),
             ws,
             );
}

function seq() {
  return function(str) {
    var args = arguments;

    if (args.length == 1) return args[0];
    else if (args.length == 2) {
      for (var betweenI = str.length; betweenI >= 0; betweenI--) {
        var firstOne = args[0](str.slice(0, betweenI));
        if (firstOne[0]) {
          var lastOne = args[1](str.slice(betweenI, str.length));
          if (lastOne[0]) {
            return [true, [firstOne[1], lastOne[1]]];
          }
        }
      }
      return [false];
    }
    else if (args.length > 2) {
      return seq(args[0], seq.apply(this, Array.prototype.slice.call(args, 1, args.length)));
    }
  }
}

function string(str0) {
  return function(str1) {
    return str0 === str1;
  }
}

