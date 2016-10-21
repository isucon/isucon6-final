package net.isucon.isucon6f;

public class Point {
    public final int id, stroke_id;
    public final double x,y;

    public Point(int id, int stroke_id, double x, double y) {
    	this.id = id;
    	this.stroke_id = stroke_id;
    	this.x = x;
    	this.y = y;
    }
}
