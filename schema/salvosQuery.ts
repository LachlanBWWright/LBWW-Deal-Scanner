import mongoose from 'mongoose';

interface salvosInterface {
    name: string,
    lastItemFound: string
}

const SalvosSchema = new mongoose.Schema<salvosInterface>({
    name: {type: String, required: true, unique: true},
    lastItemFound: {type: String, required: false, unique: false},
})

const SalvosQuery = mongoose.model<salvosInterface>('SalvosQuery', SalvosSchema);

export default SalvosQuery;