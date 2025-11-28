package dev.unifuu.grpc

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import dev.unifuu.grpc.proto.PokemonServiceGrpc
import dev.unifuu.grpc.proto.*
import io.grpc.ManagedChannelBuilder
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : ComponentActivity() {
    private lateinit var pokemonClient: PokemonClient

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Initialize gRPC client - use 10.0.2.2 for emulator, or your PC's IP for device
        pokemonClient = PokemonClient("10.0.2.2", 50051)

        setContent {
            MaterialTheme {
                PokemonScreen(pokemonClient)
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        pokemonClient.shutdown()
    }
}

@Composable
fun PokemonScreen(client: PokemonClient) {
    var query by remember { mutableStateOf("") }
    var pokemon by remember { mutableStateOf<Pokemon?>(null) }
    var errorMessage by remember { mutableStateOf("") }
    var isLoading by remember { mutableStateOf(false) }

    val scope = rememberCoroutineScope()
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(scrollState),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text(
            text = "Pokémon Finder",
            style = MaterialTheme.typography.headlineLarge,
            fontWeight = FontWeight.Bold
        )

        Text(
            text = "Enter a Pokémon name or ID (1-1025)",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        // Search input
        OutlinedTextField(
            value = query,
            onValueChange = { query = it },
            label = { Text("Pokémon Name or ID") },
            placeholder = { Text("e.g., pikachu or 25") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
            enabled = !isLoading
        )

        // Search button
        Button(
            onClick = {
                scope.launch {
                    isLoading = true
                    errorMessage = ""
                    pokemon = null

                    try {
                        val response = client.getPokemon(query)
                        if (response.success) {
                            pokemon = response.pokemon
                            errorMessage = ""
                        } else {
                            errorMessage = response.message
                            pokemon = null
                        }
                    } catch (e: Exception) {
                        errorMessage = "Error: ${e.message}"
                        pokemon = null
                    } finally {
                        isLoading = false
                    }
                }
            },
            modifier = Modifier.fillMaxWidth(),
            enabled = !isLoading && query.isNotBlank()
        ) {
            if (isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp),
                    color = MaterialTheme.colorScheme.onPrimary,
                    strokeWidth = 2.dp
                )
                Spacer(modifier = Modifier.width(8.dp))
            }
            Text(if (isLoading) "Searching..." else "Search")
        }

        // Error message
        if (errorMessage.isNotEmpty()) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.errorContainer
                )
            ) {
                Text(
                    text = errorMessage,
                    modifier = Modifier.padding(16.dp),
                    color = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        }

        // Pokemon display
        pokemon?.let { poke ->
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    // Pokemon image
                    if (poke.imageUrl.isNotEmpty()) {
                        AsyncImage(
                            model = poke.imageUrl,
                            contentDescription = poke.name,
                            modifier = Modifier
                                .size(200.dp)
                                .padding(8.dp),
                            contentScale = ContentScale.Fit
                        )
                    }

                    // Pokemon name and ID
                    Text(
                        text = poke.name,
                        style = MaterialTheme.typography.headlineMedium,
                        fontWeight = FontWeight.Bold
                    )

                    Text(
                        text = "#${String.format("%03d", poke.id)}",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )

                    // Types
                    if (poke.typesList.isNotEmpty()) {
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            poke.typesList.forEach { type ->
                                TypeBadge(type = type)
                            }
                        }
                    }

                    Divider(modifier = Modifier.padding(vertical = 8.dp))

                    // Stats
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceEvenly
                    ) {
                        StatItem(
                            label = "Height",
                            value = "${poke.height / 10.0}m"
                        )
                        StatItem(
                            label = "Weight",
                            value = "${poke.weight / 10.0}kg"
                        )
                    }
                }
            }
        }

        // Quick examples
        if (pokemon == null && errorMessage.isEmpty() && !isLoading) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.secondaryContainer
                )
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Text(
                        text = "Try these:",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold
                    )
                    Text("• pikachu")
                    Text("• charizard")
                    Text("• mewtwo")
                    Text("• 1 (Bulbasaur)")
                    Text("• 25 (Pikachu)")
                }
            }
        }
    }
}

@Composable
fun TypeBadge(type: String) {
    Surface(
        shape = RoundedCornerShape(16.dp),
        color = getTypeColor(type),
        modifier = Modifier.padding(4.dp)
    ) {
        Text(
            text = type.uppercase(),
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 6.dp),
            style = MaterialTheme.typography.labelMedium,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.surface
        )
    }
}

@Composable
fun StatItem(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            text = value,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
    }
}

@Composable
fun getTypeColor(type: String) = when (type.lowercase()) {
    "normal" -> MaterialTheme.colorScheme.surfaceVariant
    "fire" -> MaterialTheme.colorScheme.error
    "water" -> MaterialTheme.colorScheme.primary
    "electric" -> MaterialTheme.colorScheme.tertiary
    "grass" -> MaterialTheme.colorScheme.secondary
    "ice" -> MaterialTheme.colorScheme.primaryContainer
    "fighting" -> MaterialTheme.colorScheme.errorContainer
    "poison" -> MaterialTheme.colorScheme.tertiaryContainer
    "ground" -> MaterialTheme.colorScheme.secondaryContainer
    else -> MaterialTheme.colorScheme.surfaceVariant
}

class PokemonClient(host: String, port: Int) {
    private val channel = ManagedChannelBuilder
        .forAddress(host, port)
        .usePlaintext()
        .build()

    private val stub = PokemonServiceGrpc.newBlockingStub(channel)

    suspend fun getPokemon(query: String): PokemonResponse {
        return withContext(Dispatchers.IO) {
            val request = PokemonRequest.newBuilder()
                .setQuery(query)
                .build()

            stub.getPokemon(request)
        }
    }

    fun shutdown() {
        channel.shutdown()
    }
}