import { createSlice } from '@reduxjs/toolkit'

type InitialState = {
  versionNumber: string
}

const initialState: InitialState = {
  versionNumber: '',
}

export const globalSlice = createSlice({
  name: 'globalSlice',
  initialState,
  reducers: {
    setVersionNumber: (state, action) => {
      state.versionNumber = action.payload
    },
    
  },
})

export const {
  setVersionNumber,
} = globalSlice.actions

export default globalSlice.reducer
