var donParse = require('./don-parse2.js');

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

