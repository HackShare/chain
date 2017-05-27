context 'assets' do
  let(:key) { chain.mock_hsm.keys.create }
  let(:uid) { SecureRandom.uuid }

  context 'creation' do
    subject { chain.assets.create(alias: "asset-#{uid}", root_xpubs: [key.xpub], quorum: 1) }

    it 'returns the created asset' do
      expect(subject.id).not_to be_empty
    end

    it 'returns an error when required fields are missing' do
      expect { chain.assets.create(alias: :unobtanium) }.to raise_error(Chain::APIError)
    end
  end

  context 'batch creation' do
    subject {
      chain.assets.create_batch([
          {alias: "bronze-#{uid}", root_xpubs: [key.xpub], quorum: 1}, # success
          {alias: "unobtanium-#{uid}"}, # error
          {alias: "copper-#{uid}", root_xpubs: [key.xpub], quorum: 1}, #success
      ])
    }

    it 'returns successfully created assets' do
      expect(subject.successes.keys).to eq([0,2])
    end

    it 'returns errors for failed creations' do
      expect(subject.errors.keys).to eq([1])
    end

    it 'returns the reason for the error' do
      expect(subject.errors[1].code).to eq('CH202')
    end
  end

  context 'updating asset tags' do
    let(:asset1) { chain.assets.create(root_xpubs: [key.xpub], quorum: 1, tags: {x: 'one'}) }
    let(:asset2) { chain.assets.create(root_xpubs: [key.xpub], quorum: 1, tags: {y: 'one'}) }
    let(:asset3) { chain.assets.create(root_xpubs: [key.xpub], quorum: 1, tags: {z: 'one'}) }

    it 'updates individaul assets tags' do
      chain.assets.update_tags(id: asset1.id, tags: {x: 'two'})
      expect(
        chain.assets.query(filter: "id='#{asset1.id}'").first.tags
      ).to eq('x' => 'two')
    end

    it 'returns an error when no id provided' do
      expect {
        chain.assets.update_tags(tags: {x: 'three'})
      }.to raise_error(Chain::APIError)
    end

    context 'batch update' do
      subject {
        chain.assets.update_tags_batch([
          {id: asset1.id, tags: {x: 'four'}},
          {tags: {y: 'four'}},
          {id: asset2.id, tags: {y: 'four'}},
          {id: asset3.id, alias: :redundant_alias, tags: {z: 'four'}},
        ])
      }

      it 'returns successfully updated assets' do
        expect(subject.successes.keys).to eq([0,2])
      end

      it 'returns errors for failed updates' do
        expect(subject.errors.keys).to eq([1,3])
      end

      it 'returns an error for missing aliases' do
        expect(subject.errors[1].code).to eq('CH051')
      end

      it 'returns an error for redundant aliases' do
        expect(subject.errors[3].code).to eq('CH051')
      end

      it 'performs the update' do
        subject # perform batch request

        expect(
          chain.assets.query(
            filter: "id=$1 OR id=$2",
            filter_params: [asset1.id, asset2.id]
          ).all.map(&:tags).reverse
        ).to eq([
          {'x' => 'four'},
          {'y' => 'four'},
        ])
      end
    end
  end
end