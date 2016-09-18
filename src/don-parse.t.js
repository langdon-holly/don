var donParse = require('./don-parse.js');
var ps = require('./parse.js');

console.log("parse \"(1 2 3)\":");
console.log(donParse("(1 2 3)"));
console.log("[Object]:");
console.log(donParse("(1 2 3)")[0][1][1]);

console.log();

console.log("parse \"[1 2 3]\":");
console.log(donParse("[1 2 3]"));

console.log();

console.log("parse \"langdon\":");
console.log(donParse("langdon"));

console.log();

console.log("parse \"{lang\\ndon}\":");
console.log(donParse("{lang\\ndon}"));
console.log("[Object]:");
console.log(donParse("{lang\\ndon}")[0][1][1]);

console.log();

console.log("parse \"{langdon}0\":");
console.log(donParse("{langdon}0"));
console.log("[Object]:");
console.log(donParse("{langdon}0")[0][1][1]);

/*console.log();

var testList = ['a', 'r', 't', '{', 'h', 'e', 'l', 'l', 'o', '}'];
var parser = ps.seq(ps.and(donParse.expr,
                           ps.not(donParse.beginName())),
                    ps.and(donParse.exprs(),
                           ps.not(donParse.beginName())));*/
/*for (var i = 0;
     i < testList.length;
     i++) {
  if (ps.doomed(parser)) {
    console.log("doomed");
    break;}
  parser = parser.parseChar(testList[i]);
  console.log(testList[i]);}*/
/*for (;
     testList.length > 0;
     testList = testList.slice(1, testList.length)) {
  var parser = donParse.beginName();
  parser = parser.parseChar(testList[0]);
  console.log(testList[0]);
  if (ps.doomed(parser)) console.log("doomed");}*/

/*console.log();

console.log(ps.parse(ps.recurseRight(ps.or(ps.string("1+"), ps.nothing),
                                    ps.string("1")),
                     "1+1+1"));*/

