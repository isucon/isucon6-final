package net.isucon.isucon6f;

public class Token {
    public final int id;
    public final String csrf_token, created_at;

    public Token(int id, String csrf_token, String created_at) {
        this.id = id;
        this.csrf_token = csrf_token;
        this.created_at = created_at;
    }
}
