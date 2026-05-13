import "./background.css";

export function Background({ mainOpacity = 0.2 }) {
  return (
    <div aria-hidden="true" className="bg-fx-root">
      <div className="bg-fx-wash" />
      <div className="bg-mesh-stage" style={{ opacity: mainOpacity }}>
        <div className="bg-mesh-dots" />
      </div>
    </div>
  );
}
