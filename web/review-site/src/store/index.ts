import { configureStore } from "@reduxjs/toolkit";
import cardsReducer from "./slices/cardsSlice";
import viewReducer from "./slices/viewSlice";
import reviewReducer from "./slices/reviewSlice";
import commentsReducer from "./slices/commentsSlice";
import { localStorageSync } from "./middleware/localStorageSync";

export const store = configureStore({
  reducer: {
    cards: cardsReducer,
    view: viewReducer,
    review: reviewReducer,
    comments: commentsReducer,
  },
  middleware: (getDefault) => getDefault().concat(localStorageSync),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
