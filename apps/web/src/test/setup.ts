import "@testing-library/jest-dom";
import { act } from "react";

const reactModule = await import("react");
if (typeof (reactModule as Record<string, unknown>).act === "undefined") {
  Object.assign(reactModule, { act });
}
