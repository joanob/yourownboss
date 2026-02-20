# Your Own Boss

Your Own Boss es un juego web idle de gestión y producción de recursos.

Los usuarios se registran con nombre de usuario y contraseña. Dos usuarios no pueden tener el mismo nombre de usuario.

Para empezar, cada usuario crea una empresa. Esta empresa con una cantidad inicial de dinero. El dinero se muestra con tres decimales pero se almacena en base de datos como un número entero para hacer más fáciles, seguros y precisos los cálculos.

Las empresas tienen un almacenamiento de recursos iliminado. Los recursos tienen un nombre, una categoría y un precio de mercado por x cantidad. Las empresas pueden comprar y vender recursos a ese precio. 

Las empresas tienen edificios de producción. Los edificios de producción tienen diferentes procesos de producción, que utilizan unos recursos para producir otros recursos.

Una empresa puede tener varios edificios de producción del mismo tipo y producir cosas diferentes en cada uno. La producción se inicia con una acción manual, en la que se indica la cantidad de lotes que se producen. Un lote consiste en cierta cantidad de recursos de entrada que tras cierto tiempo se convertirá en cierta cantidad de recursos de salida. El mismo recurso puede ser un recurso de entrada y de salida. Al terminar ese tiempo el usuario podrá recolectar esos recursos de salida.

Los recursos y edificios de producción se crearán y modificarán desde un panel de administración.

## Ejemplos

Creo un usuario con nombre de usuario "copilot" y contraseña "1234". Creo una empresa con nombre "microsoft". Empiezo con 50,000 papelitos verdes.

Compro un invernadero por 20,000. En el invernadero puedo cultivar tomates, lechugas, fresas, ... Para cada cultivo necesito una cantidad de semillas y de agua y tras cierto tiempo recibo una cantidad de ese cultivo. Las semillas y agua que necesito la compro del mercado.

Pongo a cultivar fresas una hora. Pasada esa hora, indico que quiero recolectar los productos y empiezo a producir otra cosa.