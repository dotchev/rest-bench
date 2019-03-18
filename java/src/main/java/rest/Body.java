package rest;

import java.util.List;

public class Body {

    public static class Code {
        public int[] rgba;
        public String hex;
    }
 
    public static class Color {
        public String color;
        public String category;
        public String type;
        public Code code;
    }

    public List<Color> colors;
}
