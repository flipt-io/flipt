import { contextValidation } from './validations';

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
