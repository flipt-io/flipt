import { contextValidation, keyWithDotValidation } from './validations';

describe('contextValidation', () => {
  it('should accept valid input', () => {
    const result = contextValidation.isValidSync('{"a":"b", "c":"d"}');
    expect(result).toEqual(true);
  });
  it('should reject arrays', () => {
    let result = contextValidation.isValidSync('["a", "b"]');
    expect(result).toEqual(false);
    result = contextValidation.isValidSync('{"a":["b"]}');
    expect(result).toEqual(false);
  });
  it('should reject simple values', () => {
    let result = contextValidation.isValidSync('1');
    expect(result).toEqual(false);
    result = contextValidation.isValidSync('true');
    expect(result).toEqual(false);
  });
});

describe('keyWithDotValidation', () => {
  it('should accept key with dot', () => {
    const result = keyWithDotValidation.isValidSync('2.0');
    expect(result).toEqual(true);
  });
  it('should not accept key with invalid values', () => {
    const result = keyWithDotValidation.isValidSync('key]');
    expect(result).toEqual(false);
  });
});
