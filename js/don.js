function expr(str) {
  if (str.charAt(0) === '(') {
    return form(str);
  }
}

function form(str) {
  if (str.charAt(0) === '('
      && str.charAt(str.length-1) === ')') {
    
  }
  return [false, undefined]
}

function seq() {
  var args = arguments;
}
