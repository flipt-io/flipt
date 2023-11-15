require 'rspec'
require 'json'
require_relative '../lib/evaluation'

RSpec.describe Flipt::EvaluationClient do
  let(:client) { Flipt::EvaluationClient.new }

  describe '#variant' do
    it 'returns the variant' do
      expect(client.variant('{"key": "key1", "entity_id": "entity1"}')).to eq('{"key": "key1", "entity_id": "entity1", "variant": "variant1"}')
    end
  end
end