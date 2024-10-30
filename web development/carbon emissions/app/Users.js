import mongoose from "mongoose"

const UserSchema = new mongoose.Schema({
    name: {
        type: String,
        required: true
    },
    age: {
        type: Number, 
        required: true
    }
})

const UserModel = mongoose.model("users", UserSchema)
export default UserModel;