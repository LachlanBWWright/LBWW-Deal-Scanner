import mongoose from 'mongoose';

interface lootFarmInterface {
    name: string,
    maxPrice: number,
    minFloat: number,
    maxFloat: number,
    found: boolean
}

const LootFarmItemSchema = new mongoose.Schema<lootFarmInterface>({
    name: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 100},
    minFloat: {type: Number, required: true, min: 0, max: 1, default: 0},
    maxFloat: {type: Number, required: true, min:0, max: 1},
    found: {type: Boolean, required: true, default: false}
})

const LootFarmItem = mongoose.model<lootFarmInterface>('LootFarmItem', LootFarmItemSchema);

export default LootFarmItem;