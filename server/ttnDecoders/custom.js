/**
 * My custom logic to parse examples into decoder
 */
export function Decoder(bytes, input) {
  let data = {
    bytes: bytes,
  };

  return data
}

/**
 * My own function to parse hex string
 * Examples:
 * 4A00E92B00000000000000060E4C000000000000
 */
export function hexStringToDecoderFormat(s) {
  let arr = [];

  for (let i = 0; i < s.length; i = i + 2) {
    arr.push(s[i] + s[i + 1]);
  }

  return arr;
}
