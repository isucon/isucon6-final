package net.isucon.isucon6f;

import java.util.Arrays;

public class StrokeData {
    public final int id, room_id, width, red, green, blue;
    public final double alpha;
    private final Point[] points;
    public final String created_at;

    public StrokeData(int id, int room_id, int width, int red, int green, int blue, double alpha, Point[] points, String created_at) {
    	this.id = id;
    	this.room_id = room_id;
    	this.width = width;
    	this.red = red;
    	this.green = green;
    	this.blue = blue;
    	this.alpha = alpha;
    	this.points = Arrays.copyOf(points, points.length);
    	this.created_at = created_at;
    }
    
    public Point[] getPoints() {
    	synchronized (points) {
    		return Arrays.copyOf(points, points.length);
    	}
    }
}
