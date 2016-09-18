import React from 'react';

class Canvas extends React.Component {

  constructor(props) {
    super(props);

    this.isMouseDown = false;
  }

  getCoordinates(ev) {
    const rect = this.refs.svgElement.getBoundingClientRect();
    return {
      x: ev.clientX - rect.left,
      y: ev.clientY - rect.top,
    };
  }

  handleMouseDown(ev) {
    if (!this.isMouseDown) {
      this.isMouseDown = true;
      this.props.onStrokeStart(this.getCoordinates(ev));
    }
  }

  handleMouseMove(ev) {
    if (this.isMouseDown) {
      this.props.onStrokeMove(this.getCoordinates(ev));
    }
  }

  handleMouseUp(ev) {
    if (this.isMouseDown) {
      this.isMouseDown = false;
      this.props.onStrokeEnd(this.getCoordinates(ev));
    }
  }

  render() {
    // console.log(this.props.strokes);
    return (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        version="1.1"
        baseProfile="full"
        ref="svgElement"
        width={this.props.width}
        height={this.props.height}
        style={{
          width: this.props.width,
          height: this.props.height,
          backgroundColor: 'white',
        }}
        viewBox={`0 0 ${this.props.width} ${this.props.height}`}
        onMouseDown={(ev) => this.handleMouseDown(ev)}
        onMouseUp={(ev) => this.handleMouseUp(ev)}
        onMouseMove={(ev) => this.handleMouseMove(ev)}
      >
        {this.props.strokes
          .map((stroke) => (
            <polyline
              key={stroke.id}
              id={stroke.id}
              stroke={`rgba(${stroke.red},${stroke.green},${stroke.blue},${stroke.alpha})`}
              strokeWidth={stroke.width}
              strokeLinecap="round"
              strokeLinejoin="round"
              fill="none"
              points={stroke.points.map((point) => `${point.x},${point.y}`).join(' ')}
            />
        ))}
      </svg>
    );
  }
}

Canvas.propTypes = {
  width: React.PropTypes.number,
  height: React.PropTypes.number,
  strokes: React.PropTypes.array,
  onStrokeStart: React.PropTypes.func,
  onStrokeMove: React.PropTypes.func,
  onStrokeEnd: React.PropTypes.func,
};

export default Canvas;
