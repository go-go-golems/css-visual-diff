import { useSelector } from "react-redux";
import { selectFilteredRows } from "../store/slices/cardsSlice";
import { ReviewCard } from "./ReviewCard";

export function CardList() {
  const rows = useSelector(selectFilteredRows);

  if (rows.length === 0) {
    return (
      <div className="text-center text-text-muted py-12">
        No cards match the current filter.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {rows.map((row) => (
        <ReviewCard key={`${row.page}/${row.section}`} row={row} />
      ))}
    </div>
  );
}
