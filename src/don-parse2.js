const _ = require('lodash')
//, {asyncIterableIntoIterator, streamIntoIterator} = require('list-parsing');

const
  iterableIntoIterator
  = (to, from) =>
  { from = from[Symbol.iterator]();
    let res = to.next();
    while (!res.done) res = to.next(from.next());
    return res.value;}
, asyncIterableIntoIterator
  = async (to, from) =>
    { from = from[Symbol.asyncIterator]();
      let res = to.next();
      while (!res.done) res = to.next(await from.next());
      return res.value;}
, streamIntoIterator
  = (to, from) =>
    new Promise
    ( resolve =>
      { let res = to.next();
        const wStream
          = require('stream').Writable
            ( { write(value, encoding, cb)
                { if ((res = to.next({value})).done)
                    resolve(res.value), from.unpipe(this);
                  cb(null);}
              , final(cb)
                { resolve(to.next({done: true}).value), from.unpipe(this);
                  cb(null);}
              , objectMode: true});
        from.pipe(wStream);})
, ws = Array.from(' \t\n\r')
, delimL = Array.from('([{\\')
, delimR = Array.from(')]}|');

Object.assign
( module.exports
, { iterableIntoIterator
  , parseStream: str => streamIntoIterator(it()[Symbol.iterator](), str)
  , parseIter
    : async asyncIterable =>
      asyncIterableIntoIterator(it()[Symbol.iterator](), asyncIterable)});

function *it()
{ let value
  , index = 0
  , stack = []
  , commentLevel = 0
  , nestLevel = 0
  , line1 = 1
  , col1 = 1
  , currLine = "";

  const
    n/*ext*/
    = yielded =>
      ( value === '\n'
        ? (++line1, col1 = 1, currLine = "")
        : (++col1, currLine += value)
      , ++index
      , ({value} = yielded).done)
  , doomed
    = () =>
      ({status: 'doomed', index, result: {line1, col1, last: currLine + value}})
  , e/*of*/
    = () => ({status: 'eof', index, result: {line1, col1, last: currLine}})
  , delimited = end => (['Delimited', [...stack.pop(), end]]);

  if (n(yield)) return e();

  // shebang
  if (value === '#')
  { if (n(yield)) return e();
    if (value !== '!') return doomed();
    do if (n(yield)) return e(); while (value !== '\n');
    if (n(yield)) return e();}

  while (true)
  { if (delimL.includes(value))
      if (commentLevel--)
      { ++nestLevel;
        do
        { if (n(yield)) return e();
          if (delimL.includes(value)) ++nestLevel;
          else if (delimR.includes(value)) nestLevel--;
          else if (value === '`' && n(yield)) return e();}
        while (nestLevel);}
      else break;
    else if (value === ';') ++commentLevel;
    else if (!ws.includes(value)) return doomed();
    if (n(yield)) return e();}

  stack.push([value, []]);

  while (true)
  { if (n(yield)) return e();

    if (delimL.includes(value)) stack.push([value, []]);
    else if (delimR.includes(value))
    { if (stack.length == 1)
        return {status: 'match', index, result: delimited(value)};
      stack[stack.length - 2][1].push(delimited(value))}
    else if (value === '`')
    { if (n(yield)) return e();
      _.last(stack)[1].push(['char', value.codePointAt(0)]);}
    else
      _.last(stack)[1].push(['elem', value.codePointAt(0)])}}

