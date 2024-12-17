import { SpinnerInfinity } from "spinners-react/src/SpinnerInfinity"

export default function Loader() {
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
        width: "100%",
      }}
    >
      <SpinnerInfinity size={100} thickness={100} speed={100} color="green" />
    </div>
  )
}
