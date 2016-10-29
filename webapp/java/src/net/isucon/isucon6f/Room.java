package net.isucon.isucon6f;

import java.time.Instant;
import java.util.Arrays;
import java.util.Date;

public class Room {
    public final int id;
    public final String name;
    public final int canvas_width, canvas_height;
    public final String created_at;

    private int stroke_count, watcher_count;
    private final Stroke[] strokes; 

    public Room(int id, String name, int canvas_width, int canvas_height, Instant created_at) {
        this.id = id;
        this.name = name;
        this.canvas_width = canvas_width;
        this.canvas_height = canvas_height;
        this.created_at = Date.from(created_at).toString();
        this.strokes = new Stroke[0];
        this.stroke_count = 0;
        this.watcher_count = 0;
    }

    public synchronized void setStrokeCount(int count) {
        stroke_count = count;
    }

    public synchronized int getStrokeCount() {
        return stroke_count;
    }

    public synchronized void setWatcherCount(int count) {
        watcher_count = count;
    }

    public synchronized int getWatcherCount() {
        return watcher_count;
    }

    public Stroke[] getStrokeData() {
        synchronized (strokes) {
            return Arrays.copyOf(strokes, strokes.length);
        }
    }
}
